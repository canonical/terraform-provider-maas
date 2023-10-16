package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
	"github.com/maas/gomaasclient/entity"
)

func resourceMaasBlockDevice() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS machines' block devices.",
		CreateContext: resourceBlockDeviceCreate,
		ReadContext:   resourceBlockDeviceRead,
		UpdateContext: resourceBlockDeviceUpdate,
		DeleteContext: resourceBlockDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData,meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:BLOCK_DEVICE", d.Id())
				}
				client :=meta.(*client.Client)
				machine, err := getMachine(client, idParts[0])
				if err != nil {
					return nil, err
				}
				blockDevice, err := getBlockDevice(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":             fmt.Sprintf("%v", blockDevice.ID),
					"machine":        machine.SystemID,
					"name":           blockDevice.Name,
					"size_gigabytes": int(blockDevice.Size / (1024 * 1024 * 1024)),
					"block_size":     blockDevice.BlockSize,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The machine identifier (system ID, hostname, or FQDN) that owns the block device.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The block device name.",
			},
			"size_gigabytes": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The size of the block device (given in GB).",
			},
			"block_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     512,
				Description: "The block size of the block device. Defaults to `512`.",
			},
			"is_boot_device": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Boolean value indicating if the block device is set as the boot device.",
			},
			"partitions": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of partition resources created for the new block device. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html). And, it is computed if it's not given.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_gigabytes": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The partition size (given in GB).",
						},
						"bootable": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Boolean value indicating if the partition is set as bootable.",
						},
						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "The tags assigned to the new block device partition.",
						},
						"fs_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The file system type (e.g. `ext4`). If this is not set, the partition is unformatted.",
						},
						"label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The label assigned if the partition is formatted.",
						},
						"mount_point": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The mount point used. If this is not set, the partition is not mounted. This is used only the partition is formatted.",
						},
						"mount_options": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The options used for the partition mount.",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the partition.",
						},
					},
				},
			},
			"model": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				RequiredWith:  []string{"serial"},
				ConflictsWith: []string{"id_path"},
				AtLeastOneOf:  []string{"model", "id_path"},
				Description:   "Model of the block device. Used in conjunction with `serial` argument. Conflicts with `id_path`. This argument is computed if it's not given.",
			},
			"serial": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				RequiredWith:  []string{"model"},
				ConflictsWith: []string{"id_path"},
				Description:   "Serial number of the block device. Used in conjunction with `model` argument. Conflicts with `id_path`. This argument is computed if it's not given.",
			},
			"id_path": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"model", "serial"},
				AtLeastOneOf:  []string{"model", "id_path"},
				Description:   "Only used if `model` and `serial` cannot be provided. This should be a path that is fixed and doesn't change depending on the boot order or kernel version. This argument is computed if it's not given.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "A set of tag names assigned to the new block device. This argument is computed if it's not given.",
				},
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Block device UUID.",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Block device path.",
			},
		},
	}
}

func resourceBlockDeviceCreate(ctx context.Context, d *schema.ResourceData,meta interface{}) diag.Diagnostics {
	client :=meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := findBlockDevice(client, machine.SystemID, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if blockDevice == nil {
		blockDevice, err = client.BlockDevices.Create(machine.SystemID, getBlockDeviceParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%v", blockDevice.ID))

	return resourceBlockDeviceUpdate(ctx, d,meta)
}

func resourceBlockDeviceRead(ctx context.Context, d *schema.ResourceData,meta interface{}) diag.Diagnostics {
	client :=meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := client.BlockDevice.Get(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"partitions": getBlockDevicePartitionsTFState(blockDevice),
		"model":      blockDevice.Model,
		"serial":     blockDevice.Serial,
		"id_path":    blockDevice.IDPath,
		"tags":       blockDevice.Tags,
		"uuid":       blockDevice.UUID,
		"path":       blockDevice.Path,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBlockDeviceUpdate(ctx context.Context, d *schema.ResourceData,meta interface{}) diag.Diagnostics {
	client :=meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := client.BlockDevice.Update(machine.SystemID, id, getBlockDeviceParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err := setBlockDeviceTags(client, d, blockDevice); err != nil {
		return diag.FromErr(err)
	}
	if p, ok := d.GetOk("is_boot_device"); ok && p.(bool) {
		if err := client.BlockDevice.SetBootDisk(machine.SystemID, id); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := updateBlockDevicePartitions(client, d, blockDevice); err != nil {
		return diag.FromErr(err)
	}

	return resourceBlockDeviceRead(ctx, d,meta)
}

func resourceBlockDeviceDelete(ctx context.Context, d *schema.ResourceData,meta interface{}) diag.Diagnostics {
	client :=meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.BlockDevice.Delete(machine.SystemID, id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getBlockDeviceParams(d *schema.ResourceData) *entity.BlockDeviceParams {
	return &entity.BlockDeviceParams{
		Name:      d.Get("name").(string),
		Size:      d.Get("size_gigabytes").(int) * 1024 * 1024 * 1024,
		BlockSize: d.Get("block_size").(int),
		Model:     d.Get("model").(string),
		Serial:    d.Get("serial").(string),
		IDPath:    d.Get("id_path").(string),
	}
}

func findBlockDevice(client *client.Client, machineID string, identifier string) (*entity.BlockDevice, error) {
	blockDevices, err := client.BlockDevices.Get(machineID)
	if err != nil {
		return nil, err
	}
	for _, b := range blockDevices {
		if fmt.Sprintf("%v", b.ID) == identifier || b.Name == identifier || b.IDPath == identifier || b.Path == identifier {
			return &b, nil
		}
	}
	return nil, nil
}

func getBlockDevice(client *client.Client, machineID string, identifier string) (*entity.BlockDevice, error) {
	blockDevice, err := findBlockDevice(client, machineID, identifier)
	if err != nil {
		return nil, err
	}
	if blockDevice == nil {
		return nil, fmt.Errorf("block device (%s) was not found on machine (%s)", identifier, machineID)
	}
	return blockDevice, nil
}

func setBlockDeviceTags(client *client.Client, d *schema.ResourceData, blockDevice *entity.BlockDevice) error {
	p, ok := d.GetOk("tags")
	if !ok {
		return nil
	}
	tags := p.(*schema.Set).List()
	blockDevice, err := client.BlockDevice.Get(blockDevice.SystemID, blockDevice.ID)
	if err != nil {
		return err
	}
	// Remove existing tags
	for _, t := range blockDevice.Tags {
		if _, err = client.BlockDevice.RemoveTag(blockDevice.SystemID, blockDevice.ID, t); err != nil {
			return err
		}
	}
	// Add new tags
	for _, t := range tags {
		if _, err = client.BlockDevice.AddTag(blockDevice.SystemID, blockDevice.ID, t.(string)); err != nil {
			return err
		}
	}
	return nil
}

func getBlockDevicePartitionsTFState(blockDevice *entity.BlockDevice) []map[string]interface{} {
	partitions := make([]map[string]interface{}, len(blockDevice.Partitions))
	for i, p := range blockDevice.Partitions {
		part := map[string]interface{}{
			"size_gigabytes": int(p.Size / (1024 * 1024 * 1024)),
			"bootable":       p.Bootable,
			"tags":           p.Tags,
			"fs_type":        p.FileSystem.FSType,
			"label":          p.FileSystem.Label,
			"mount_point":    p.FileSystem.MountPoint,
			"mount_options":  p.FileSystem.MountOptions,
			"path":           p.Path,
		}
		partitions[i] = part
	}
	return partitions
}

func updateBlockDevicePartitions(client *client.Client, d *schema.ResourceData, blockDevice *entity.BlockDevice) error {
	p, ok := d.GetOk("partitions")
	if !ok {
		return nil
	}
	// Remove existing partitions
	for _, part := range blockDevice.Partitions {
		if err := client.BlockDevicePartition.Delete(blockDevice.SystemID, blockDevice.ID, part.ID); err != nil {
			return err
		}

	}
	// Create new partitions given by the user
	partitions := p.([]interface{})
	for _, part := range partitions {
		partition := part.(map[string]interface{})
		partitionParams := entity.BlockDevicePartitionParams{
			Size:     partition["size_gigabytes"].(int) * 1024 * 1024 * 1024,
			Bootable: partition["bootable"].(bool),
		}
		blockDevicePartition, err := client.BlockDevicePartitions.Create(blockDevice.SystemID, blockDevice.ID, &partitionParams)
		if err != nil {
			return err
		}
		tags := partition["tags"].(*schema.Set).List()
		for _, t := range tags {
			if _, err := client.BlockDevicePartition.AddTag(blockDevice.SystemID, blockDevice.ID, blockDevicePartition.ID, t.(string)); err != nil {
				return err
			}
		}
		if fsType := partition["fs_type"].(string); fsType != "" {
			label := partition["label"].(string)
			if _, err := client.BlockDevicePartition.Format(blockDevice.SystemID, blockDevice.ID, blockDevicePartition.ID, fsType, label); err != nil {
				return err
			}
			if mountPoint := partition["mount_point"].(string); mountPoint != "" {
				mountOptions := partition["mount_options"].(string)
				if _, err := client.BlockDevicePartition.Mount(blockDevice.SystemID, blockDevice.ID, blockDevicePartition.ID, mountPoint, mountOptions); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
