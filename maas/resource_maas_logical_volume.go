package maas

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASLogicalVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLogicalVolumeCreate,
		ReadContext:   resourceLogicalVolumeRead,
		UpdateContext: resourceLogicalVolumeUpdate,
		DeleteContext: resourceLogicalVolumeDelete,

		Schema: map[string]*schema.Schema{
			"fs_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The file system type (e.g. `ext4`). If this is not set, the volume is unformatted.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The machine identifier (system ID, hostname, or FQDN) that owns the volume group.",
			},
			"mount_options": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comma separated options used for the logical volume mount.",
			},
			"mount_point": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The mount point used. If this is not set, the logical volume is not mounted.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name for this logical volume",
			},
			"size_gigabytes": {
				Type:        schema.TypeInt,
				Required:    true,
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

func resourceLogicalVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Validate the file system and mounting information before attempting to create the logical volume
	// If a mount point is specified, then fs_type is required
	if mountPoint := d.Get("mount_point").(string); mountPoint != "" {
		if d.Get("fs_type").(string) == "" {
			return diag.Errorf("invalid block device mount configuration: fs_type must be specified when mount_point is set")
		}
	}

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
		Size: int64(d.Get("size_gigabytes").(int)) * GigaBytes,
	}

	createdLVM, err := client.VolumeGroup.CreateLogicalVolume(machine.SystemID, volumeGroup.ID, LVMParams)
	if err != nil {
		return diag.FromErr(err)
	}

	formattedDevice, err := formatAndMountVirtualBlockDevice(client, createdLVM, d)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", formattedDevice.ID))

	return resourceLogicalVolumeRead(ctx, d, meta)
}

func resourceLogicalVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceLogicalVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	params := entity.BlockDeviceParams{
		Name: d.Get("name").(string),
		Size: int64(d.Get("size_gigabytes").(int)) * GigaBytes,
	}

	updatedLVM, err := client.BlockDevice.Update(machine.SystemID, id, &params)
	if err != nil {
		return diag.FromErr(err)
	}

	formattedDevice, err := formatAndMountVirtualBlockDevice(client, updatedLVM, d)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", formattedDevice.ID))

	return resourceLogicalVolumeRead(ctx, d, meta)
}

func resourceLogicalVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	name = strings.ReplaceAll(name, fmt.Sprintf("%s-", volumeGroup.Name), "")

	tfState := map[string]interface{}{
		"fs_type":        logicalVolume.Filesystem.FSType,
		"machine":        machine.SystemID,
		"mount_options":  logicalVolume.Filesystem.MountOptions,
		"mount_point":    logicalVolume.Filesystem.MountPoint,
		"name":           name,
		"size_gigabytes": math.Round(float64(logicalVolume.Size) / GigaBytes),
		"volume_group":   fmt.Sprintf("%v", volumeGroup.ID),
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.Errorf("Could not set logical volume state: %v", err)
	}

	return nil
}

func formatAndMountVirtualBlockDevice(client *client.Client, virtualBlockDevice *entity.BlockDevice, d *schema.ResourceData) (*entity.BlockDevice, error) {
	var err error

	// Format the device if fs_type is specified
	if fsType := d.Get("fs_type").(string); fsType != "" {
		virtualBlockDevice, err = client.BlockDevice.Format(virtualBlockDevice.SystemID, virtualBlockDevice.ID, fsType)
		if err != nil {
			return nil, fmt.Errorf("failed to format block device: %w", err)
		}
	}

	// Mount the device if mount_point is specified
	if mountPoint := d.Get("mount_point").(string); mountPoint != "" {
		mountOptions := d.Get("mount_options").(string)

		virtualBlockDevice, err = client.BlockDevice.Mount(virtualBlockDevice.SystemID, virtualBlockDevice.ID, mountPoint, mountOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to mount block device: %w", err)
		}
	}

	return virtualBlockDevice, nil
}
