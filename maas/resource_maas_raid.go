package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASRAID() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"block_devices": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "The list of block devices to be included in the RAID.\n*Note*: For the boot disk, a partition should be supplied instead, as MAAS would otherwise automatically create one.",
				AtLeastOneOf: []string{"block_devices", "partitions"},
			},
			"fs_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The file system type (e.g. `ext4`). If this is not set, the RAID is unformatted.",
			},
			"level": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The RAID Level. Valid levels are: `0, 1, 5, 6, 10`",
				ValidateFunc: validation.IntInSlice(
					[]int{0, 1, 5, 6, 10},
				),
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The machine identifier (system ID, hostname, or FQDN) that owns the RAID.",
			},
			"mount_options": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comma seperated options used for the RAID mount.",
			},
			"mount_point": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The mount point used. If this is not set, the RAID is not mounted.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name for the RAID",
			},
			"partitions": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "The list of partitions to be included in the RAID.",
				AtLeastOneOf: []string{"block_devices", "partitions"},
			},
			"size_gigabytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The volume size (given in GB).",
			},
			"spare_devices": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of spare block devices for the RAID.",
			},
			"spare_partitions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of spare partitions for the RAID.",
			},
		},

		CustomizeDiff: schema.CustomizeDiffFunc(func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			devices := d.Get("block_devices").(*schema.Set)
			partitions := d.Get("partitions").(*schema.Set)
			spareDevices := d.Get("spare_devices").(*schema.Set)
			sparePartitions := d.Get("spare_partitions").(*schema.Set)

			if devices.Len() == 0 && spareDevices.Len() > 0 {
				return fmt.Errorf("`spare_devices` cannot be specified unless `block_devices` is also specified")
			}
			if partitions.Len() == 0 && sparePartitions.Len() > 0 {
				return fmt.Errorf("`spare_partitions` cannot be specified unless `partitions` is also specified")
			}
			return nil
		}),
	}
}
