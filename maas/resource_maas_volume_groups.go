package maas

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasVolumeGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS Volume Groups.",
		CreateContext: resourceMaasVolumeGroupCreate,
		ReadContext:   resourceMaasVolumeGroupRead,
		UpdateContext: resourceMaasVolumeGroupUpdate,
		DeleteContext: resourceMaasVolumeGroupDelete,

		Schema: map[string]*schema.Schema{
			"available_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The volume group available size (B).",
			},
			"block_devices": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of block device names to be included in this volume group.",
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

func resourceMaasVolumeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	volumeGroupParams := entity.VolumeGroupCreateParams{
		Name:         d.Get("name").(string),
		BlockDevices: d.Get("block_devices").([]string),
	}

	volumeGroup, err := client.VolumeGroups.Create(machine.SystemID, &volumeGroupParams)
	if err != nil {
		return diag.Errorf("Could not create volume group: %v", err)
	}

	d.SetId((fmt.Sprintf("%v", volumeGroup.ID)))

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

	var blockDevices []string
	for _, blockDevice := range volumeGroup.LogicalVolumes {
		blockDevices = append(blockDevices, blockDevice.Name)
	}

	tfstate := map[string]interface{}{
		"block_devices":  blockDevices,
		"machine":        volumeGroup.SystemID,
		"name":           volumeGroup.Name,
		"size":           volumeGroup.Size,
		"used_size":      volumeGroup.UsedSize,
		"available_size": volumeGroup.AvailableSize,
		"uuid":           volumeGroup.UUID,
	}
	if err := setTerraformState(d, tfstate); err != nil {
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
	if d.HasChange("block_Devices") {
		oldBlockDevices, newBlockDevices := d.GetChange("block_devices")

		oldDeviceList := oldBlockDevices.([]string)
		newDeviceList := newBlockDevices.([]string)

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

	d.SetId((fmt.Sprintf("%v", volumeGroup.ID)))

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
