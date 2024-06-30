package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
)

func dataSourceMaasBlockDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext:   dataSourceBlockDeviceRead,
		Schema: map[string]*schema.Schema{
			"block_size": {
				Type:        schema.TypeInt,
				Computed:      true,
				Description: "The block size of the block device. Defaults to `512`.",
			},
			"id_path": {
				Type:          schema.TypeString,
				Computed:      true,
				Description:   "Only used if `model` and `serial` cannot be provided. This should be a path that is fixed and doesn't change depending on the boot order or kernel version. This argument is computed if it's not given.",
			},
			"is_boot_device": {
				Type:        schema.TypeBool,
				Computed:      true,
				Description: "Boolean value indicating if the block device is set as the boot device.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The machine identifier (system ID, hostname, or FQDN) that owns the block device.",
			},
			"model": {
				Type:          schema.TypeString,
				Computed:      true,
				Description:   "Model of the block device. Used in conjunction with `serial` argument. Conflicts with `id_path`. This argument is computed if it's not given.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The block device name.",
			},
			"partitions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of partition resources created for the new block device. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html). And, it is computed if it's not given.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bootable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Boolean value indicating if the partition is set as bootable.",
						},
						"fs_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The file system type (e.g. `ext4`). If this is not set, the partition is unformatted.",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The label assigned if the partition is formatted.",
						},
						"mount_options": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The options used for the partition mount.",
						},
						"mount_point": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The mount point used. If this is not set, the partition is not mounted. This is used only the partition is formatted.",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the partition.",
						},
						"size_gigabytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The partition size (given in GB).",
						},
						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "The tags assigned to the new block device partition.",
						},
					},
				},
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Block device path.",
			},
			"serial": {
				Type:          schema.TypeString,
				Computed:      true,
				Description:   "Serial number of the block device. Used in conjunction with `model` argument. Conflicts with `id_path`. This argument is computed if it's not given.",
			},
			"size_gigabytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the block device (given in GB).",
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names assigned to the new block device. This argument is computed if it's not given.",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Block device UUID.",
			},
		},
	}
}

func dataSourceBlockDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := findBlockDevice(client, machine.SystemID, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"id":             fmt.Sprintf("%v", blockDevice.ID),
		"partitions":     getBlockDevicePartitionsTFState(blockDevice),
		"model":          blockDevice.Model,
		"serial":         blockDevice.Serial,
		"id_path":        blockDevice.IDPath,
		"tags":           blockDevice.Tags,
		"uuid":           blockDevice.UUID,
		"path":           blockDevice.Path,
		"machine":        machine.SystemID,
		"name":           blockDevice.Name,
		"size_gigabytes": int(blockDevice.Size / (1024 * 1024 * 1024)),
		"block_size":     blockDevice.BlockSize,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
