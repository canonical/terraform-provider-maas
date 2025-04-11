package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASLogicalVolumeLvm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMAASLogicalVolumeLvmCreate,
		DeleteContext: resourceMAASLogicalVolumeLvmDelete,
		ReadContext:   resourceMAASLogicalVolumeLvmRead,

		Schema: map[string]*schema.Schema{
			"fs_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The file system type (e.g. `ext4`). If this is not set, the volume is unformatted.",
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
				ForceNew:    true,
				Description: "The name for this logical volume",
			},
			"size_gigabytes": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The volume size (given in GB).",
			},
			"volume_group": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The volume group identifier (ID or name) to apply this logical volume on top of.",
			},
		},
	}
}

func resourceMAASLogicalVolumeLvmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	volumeGroup, err := getVolumeGroup(client, machine.SystemID, d.Get("volume_group").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	LVMParams := &entity.LogicalVolumeParams{
		Name: d.Get("name").(string),
		Size: int64(d.Get("size_gigabytes").(int)) * 1024 * 1024 * 1024,
	}

	createdLvmBlockDevice, err := client.VolumeGroup.CreateLogicalVolume(machine.SystemID, volumeGroup.ID, LVMParams)
	if err != nil {
		return diag.FromErr(err)
	}

	formattedDevice, err := client.BlockDevice.Format(machine.SystemID, createdLvmBlockDevice.ID, d.Get("fs_type").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId((fmt.Sprintf("%v", formattedDevice.ID)))

	return resourceMAASLogicalVolumeLvmRead(ctx, d, meta)
}

func resourceMAASLogicalVolumeLvmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	volumeGroup, err := getVolumeGroup(client, machine.SystemID, d.Get("volume_group").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.VolumeGroup.DeleteLogicalVolume(machine.SystemID, volumeGroup.ID, id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMAASLogicalVolumeLvmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	volumeGroup, err := getVolumeGroup(client, machine.SystemID, d.Get("volume_group").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// logical volumes are technically block devices
	logicalVolume, err := client.BlockDevice.Get(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// this has the format VG name-BD Name, we only want BD Name
	name := logicalVolume.Name
	name = strings.Replace(name, fmt.Sprintf("%s-", volumeGroup.Name), "", -1)

	tfState := map[string]interface{}{
		"fs_type":        logicalVolume.Filesystem.FSType,
		"machine":        machine.SystemID,
		"name":           name,
		"size_gigabytes": (logicalVolume.Size / (1024 * 1024 * 1024)),
		"volume_group":   fmt.Sprintf("%v", volumeGroup.ID),
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.Errorf("Could not set logical volume state: %v", err)
	}

	return nil
}
