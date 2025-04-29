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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASRAID() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRAIDCreate,
		ReadContext:   resourceRAIDRead,
		UpdateContext: resourceRAIDUpdate,
		DeleteContext: resourceRAIDDelete,

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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The RAID Level. Valid levels are: `\"0\", \"1\", \"5\", \"6\", \"10\"`",
				ValidateFunc: validation.StringInSlice(
					[]string{"0", "1", "5", "6", "10"},
					false,
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

func resourceRAIDCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	blockDevices := convertToStringSlice(d.Get("block_devices").(*schema.Set).List())
	partitions := convertToStringSlice(d.Get("partitions").(*schema.Set).List())

	spareDevices := convertToStringSlice(d.Get("spare_devices").(*schema.Set).List())
	sparePartitions := convertToStringSlice(d.Get("spare_partitions").(*schema.Set).List())

	// MAAS has an unelpful error if you supply a block device that has partitions, so
	// perform the check and turn it into a more helpful error
	if err = verifyRAIDDevices(client, machine.SystemID, blockDevices); err != nil {
		return diag.FromErr(err)
	}

	if err = verifyRAIDDevices(client, machine.SystemID, spareDevices); err != nil {
		return diag.FromErr(err)
	}

	// validation on raid level vs disk count
	deviceCount := len(blockDevices) + len(partitions)
	level := d.Get("level").(string)

	if deviceCount <= 1 {
		return diag.Errorf("RAIDs require at least two disks")
	}

	if (level == "5" || level == "10") && deviceCount < 3 {
		return diag.Errorf("RAID level %v requires at least three disks", level)
	}

	if level == "6" && deviceCount < 4 {
		return diag.Errorf("RAID level %v requires at least four disks", level)
	}

	RAIDParams := &entity.RAIDCreateParams{
		Name:            d.Get("name").(string),
		Level:           fmt.Sprintf("raid-%v", level),
		BlockDevices:    blockDevices,
		Partitions:      partitions,
		SpareDevices:    spareDevices,
		SparePartitions: sparePartitions,
	}

	raid, err := client.RAIDs.Create(machine.SystemID, RAIDParams)
	if err != nil {
		return diag.Errorf("Could not create RAID: %v", err)
	}

	d.SetId(fmt.Sprintf("%v", raid.ID))

	return nil
}

func resourceRAIDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceRAIDUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceRAIDDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.RAID.Delete(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func verifyRAIDDevices(client *client.Client, machineID string, devices []string) error {
	// ensure no block devices with partitions have been passed to the RAID
	blockDevices, err := client.BlockDevices.Get(machineID)
	if err != nil {
		return err
	}

	for _, blockDevice := range blockDevices {
		id := fmt.Sprintf("%d", blockDevice.ID)
		if slices.Contains(devices, id) && len(blockDevice.Partitions) > 0 {
			return fmt.Errorf("cannot create a RAID from a block device with partitions, supply the partitions for %v instead", blockDevice.Name)
		}
	}

	return nil
}
