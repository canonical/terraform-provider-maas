package maas

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASVolumeGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS Volume Groups, and construct them from partion-less block devices.",
		CreateContext: resourceMAASVolumeGroupCreate,
		ReadContext:   resourceMAASVolumeGroupRead,
		UpdateContext: resourceMAASVolumeGroupUpdate,
		DeleteContext: resourceMAASVolumeGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMAASVolumeGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"block_devices": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "The list of block device ids to be included in this volume group.\n*Note*: For the boot disk, a partition should be supplied instead, as MAAS would otherwise automatically create one.",
				AtLeastOneOf: []string{"block_devices", "partitions"},
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The machine identifier (system ID, hostname, or FQDN) that owns the volume group.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name for this volume group",
			},
			"partitions": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "The list of partition ids to be included in this volume group.",
				AtLeastOneOf: []string{"block_devices", "partitions"},
			},
			"size_gigabytes": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The volume group size (GiB).",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Volume group UUID.",
			},
		},
	}
}

func resourceMAASVolumeGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE_ID/VOLUME_GROUP_ID", d.Id())
	}

	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, idParts[0])
	if err != nil {
		return nil, err
	}

	// we have a dependency on the specific machine in the read function
	d.Set("machine", machine.SystemID)

	volumeGroup, err := getVolumeGroup(client, machine.SystemID, idParts[1])
	if err != nil {
		return nil, err
	}

	d.SetId(fmt.Sprintf("%v", volumeGroup.ID))

	return []*schema.ResourceData{d}, nil
}

func resourceMAASVolumeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	blockDevices := convertToStringSlice(d.Get("block_devices").(*schema.Set).List())
	partitions := convertToStringSlice(d.Get("partitions").(*schema.Set).List())

	bootDisk := fmt.Sprintf("%v", machine.BootDisk.ID)
	if slices.Contains(blockDevices, bootDisk) {
		return diag.Errorf("Cannot add the boot disk %v (%v) as a block device, provide partitions on top of it instead.", bootDisk, machine.BootDisk.Name)
	}

	volumeGroupParams := entity.VolumeGroupCreateParams{
		Name:         d.Get("name").(string),
		BlockDevices: blockDevices,
		Partitions:   partitions,
	}

	volumeGroup, err := client.VolumeGroups.Create(machine.SystemID, &volumeGroupParams)
	if err != nil {
		return diag.Errorf("Could not create volume group: %v", err)
	}

	d.SetId(fmt.Sprintf("%v", volumeGroup.ID))

	return resourceMAASVolumeGroupRead(ctx, d, meta)
}

func resourceMAASVolumeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	volumeGroup, err := client.VolumeGroup.Get(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	blockDevices, partitions := findVolumeGroupDevices(volumeGroup)

	tfState := map[string]interface{}{
		"block_devices":  blockDevices,
		"machine":        volumeGroup.SystemID,
		"name":           volumeGroup.Name,
		"partitions":     partitions,
		"size_gigabytes": int64(volumeGroup.Size / (1024 * 1024 * 1024)),
		"uuid":           volumeGroup.UUID,
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.Errorf("Could not set volume group state: %v", err)
	}

	return nil
}

func resourceMAASVolumeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var addBlockDevices []string

	var removeBlockDevices []string

	if d.HasChange("block_devices") {
		oldBlockDevices, newBlockDevices := d.GetChange("block_devices")

		oldDeviceList := convertToStringSlice(oldBlockDevices.(*schema.Set).List())
		newDeviceList := convertToStringSlice(newBlockDevices.(*schema.Set).List())

		for _, device := range newDeviceList {
			if !slices.Contains(oldDeviceList, device) {
				addBlockDevices = append(addBlockDevices, device)
			}
		}

		for _, device := range oldDeviceList {
			if !slices.Contains(newDeviceList, device) {
				removeBlockDevices = append(removeBlockDevices, device)
			}
		}
	}

	var addPartitions []string

	var removePartitions []string

	if d.HasChange("partitions") {
		oldPartitions, newPartitions := d.GetChange("partitions")

		oldPartitionList := convertToStringSlice(oldPartitions.(*schema.Set).List())
		newPartitionList := convertToStringSlice(newPartitions.(*schema.Set).List())

		for _, partition := range newPartitionList {
			if !slices.Contains(oldPartitionList, partition) {
				addPartitions = append(addPartitions, partition)
			}
		}

		for _, partition := range oldPartitionList {
			if !slices.Contains(newPartitionList, partition) {
				removePartitions = append(removePartitions, partition)
			}
		}
	}

	updateParams := entity.VolumeGroupUpdateParams{
		Name:               d.Get("name").(string),
		AddBlockDevices:    addBlockDevices,
		RemoveBlockDevices: removeBlockDevices,
		AddPartitions:      addPartitions,
		RemovePartitions:   removePartitions,
	}

	volumeGroup, err := client.VolumeGroup.Update(machine.SystemID, id, &updateParams)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", volumeGroup.ID))

	return resourceMAASVolumeGroupRead(ctx, d, meta)
}

func resourceMAASVolumeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.VolumeGroup.Delete(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func findVolumeGroupDevices(volumeGroup *entity.VolumeGroup) ([]string, []string) {
	var blockDevices []string

	var partitions []string

	for _, device := range volumeGroup.Devices.([]interface{}) {
		thisDevice := device.(map[string]interface{})

		if id, ok := thisDevice["id"]; ok {
			id := fmt.Sprintf("%v", id)

			// partitions have a parent device_id, block devices do not
			if _, ok := thisDevice["device_id"]; ok {
				partitions = append(partitions, id)
			} else {
				blockDevices = append(blockDevices, id)
			}
		}
	}

	return blockDevices, partitions
}

func getVolumeGroup(client *client.Client, machineID string, identifier string) (*entity.VolumeGroup, error) {
	volumegroups, err := client.VolumeGroups.Get(machineID)
	if err != nil {
		return nil, err
	}

	if volumegroups == nil {
		return nil, fmt.Errorf("volume group %v was not found on machine %v", identifier, machineID)
	}

	for _, vg := range volumegroups {
		if fmt.Sprintf("%v", vg.ID) == identifier || vg.Name == identifier {
			return &vg, nil
		}
	}

	return nil, err
}
