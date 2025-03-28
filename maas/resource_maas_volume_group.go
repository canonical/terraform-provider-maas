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

func resourceMaasVolumeGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS Volume Groups, and construct them from partion-less block devices.",
		CreateContext: resourceMaasVolumeGroupCreate,
		ReadContext:   resourceMaasVolumeGroupRead,
		UpdateContext: resourceMaasVolumeGroupUpdate,
		DeleteContext: resourceMaasVolumeGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMaasVolumeGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"available_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The volume group available size (B).",
			},
			"block_devices": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of block device ids to be included in this volume group.",
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
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The volume group size (B).",
			},
			"used_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The volume group used size (B).",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Volume group UUID.",
			},
		},
	}
}

func resourceMaasVolumeGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE_ID:VOLUME_GROUP_ID", d.Id())
	}

	client := meta.(*ClientConfig).Client
	machine, err := getMachine(client, idParts[0])
	if err != nil {
		return nil, err
	}

	volumeGroup, err := getVolumeGroup(client, machine.SystemID, idParts[1])
	if err != nil {
		return nil, err
	}

	blockDevices := findVolumeGroupBlockDevices(volumeGroup)

	tfState := map[string]interface{}{
		"block_devices":  blockDevices,
		"machine":        volumeGroup.SystemID,
		"name":           volumeGroup.Name,
		"size":           volumeGroup.Size,
		"used_size":      volumeGroup.UsedSize,
		"available_size": volumeGroup.AvailableSize,
		"uuid":           volumeGroup.UUID,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return nil, err
	}

	d.SetId(fmt.Sprintf("%v", volumeGroup.ID))

	return []*schema.ResourceData{d}, nil
}

func resourceMaasVolumeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	block_devices := d.Get("block_devices").(*schema.Set).List()

	volumeGroupParams := entity.VolumeGroupCreateParams{
		Name:         d.Get("name").(string),
		BlockDevices: convertToStringSlice(block_devices),
	}

	volumeGroup, err := client.VolumeGroups.Create(machine.SystemID, &volumeGroupParams)
	if err != nil {
		return diag.Errorf("Could not create volume group: %v", err)
	}

	d.SetId(fmt.Sprintf("%v", volumeGroup.ID))

	return resourceMaasVolumeGroupRead(ctx, d, meta)
}

func resourceMaasVolumeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	blockDevices := findVolumeGroupBlockDevices(volumeGroup)

	tfState := map[string]interface{}{
		"block_devices":  blockDevices,
		"machine":        volumeGroup.SystemID,
		"name":           volumeGroup.Name,
		"size":           volumeGroup.Size,
		"used_size":      volumeGroup.UsedSize,
		"available_size": volumeGroup.AvailableSize,
		"uuid":           volumeGroup.UUID,
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.Errorf("Could not set volume group state: %v", err)
	}

	return nil
}

func resourceMaasVolumeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	updateParams := entity.VolumeGroupUpdateParams{
		Name:               d.Get("name").(string),
		AddBlockDevices:    addBlockDevices,
		RemoveBlockDevices: removeBlockDevices,
	}

	volumeGroup, err := client.VolumeGroup.Update(machine.SystemID, id, &updateParams)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", volumeGroup.ID))

	return resourceMaasVolumeGroupRead(ctx, d, meta)
}

func resourceMaasVolumeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func findVolumeGroupBlockDevices(volumeGroup *entity.VolumeGroup) []string {
	var blockDevices []string
	for _, device := range volumeGroup.Devices.([]interface{}) {
		thisDevice := device.(map[string]interface{})

		var deviceId string

		// partitions list a device id of the parent block device
		if did, ok := thisDevice["device_id"]; ok {
			deviceId = fmt.Sprintf("%v", did)
		} else if id, ok := thisDevice["id"]; ok {
			deviceId = fmt.Sprintf("%v", id)
		} else {
			continue
		}

		blockDevices = append(blockDevices, deviceId)
	}
	return blockDevices
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
	return nil, nil
}
