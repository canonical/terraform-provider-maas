package maas

import (
	"context"
	"fmt"
	"log"
	"math"
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
		Description:   "Provides a resource to manage MAAS RAIDs, and construct them from block devices and partitions.",
		CreateContext: resourceRAIDCreate,
		ReadContext:   resourceRAIDRead,
		UpdateContext: resourceRAIDUpdate,
		DeleteContext: resourceRAIDDelete,

		Schema: map[string]*schema.Schema{
			"block_devices": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "The list of block devices to be included in the RAID.\n*Note*: The boot disk cannot participate in the RAID as a block device, a partition on top of it should be supplied instead.\n*Note*: Block devices with partitions are not valid targets to construct a RAID, supply their partitions instead.",
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
				Description: "Comma separated options used for the RAID mount.",
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
				Description: "The list of spare block devices for the RAID.\n*Note*: The boot disk cannot participate in the RAID as a block device, a partition on top of it should be supplied instead.\n*Note*: Block devices with partitions are not valid targets to construct a RAID, supply their partitions instead.",
			},
			"spare_partitions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of spare partitions for the RAID.",
			},
		},
	}
}

func resourceRAIDCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Validate the file system and mounting information before attempting to create the RAID
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

	// Check the RAID configuration is valid
	if err = verifyRAIDConfig(client, machine, d); err != nil {
		return diag.FromErr(err)
	}

	// We *should* have a valid configuration to be able to create the RAID
	createRAIDParams := &entity.RAIDCreateParams{
		Name:            d.Get("name").(string),
		Level:           fmt.Sprintf("raid-%v", d.Get("level").(string)),
		BlockDevices:    convertToStringSlice(d.Get("block_devices").(*schema.Set).List()),
		Partitions:      convertToStringSlice(d.Get("partitions").(*schema.Set).List()),
		SpareDevices:    convertToStringSlice(d.Get("spare_devices").(*schema.Set).List()),
		SparePartitions: convertToStringSlice(d.Get("spare_partitions").(*schema.Set).List()),
	}

	raid, err := client.RAIDs.Create(machine.SystemID, createRAIDParams)
	if err != nil {
		return diag.Errorf("Could not create RAID: %v", err)
	}

	// We perform mounting and formatting operations on the virtual block device, rather than a part of the RAID creation
	if _, err = formatAndMountVirtualBlockDevice(client, &raid.VirtualDevice, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", raid.ID))

	return resourceRAIDRead(ctx, d, meta)
}

func resourceRAIDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	raid, err := client.RAID.Get(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// block devices and partitions are stored on the same object, we need to split them out
	devices, partitions, err := splitDeviceTypes(raid.Devices)
	if err != nil {
		return diag.FromErr(err)
	}

	spareDevices, sparePartitions, err := splitDeviceTypes(raid.SpareDevices)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update the Terraform state
	tfstate := map[string]interface{}{
		"block_devices":    devices,
		"fs_type":          raid.VirtualDevice.Filesystem.FSType,
		"level":            strings.ReplaceAll(raid.Level, "raid-", ""),
		"machine":          raid.SystemID,
		"mount_options":    raid.VirtualDevice.Filesystem.MountOptions,
		"mount_point":      raid.VirtualDevice.Filesystem.MountPoint,
		"name":             raid.Name,
		"partitions":       partitions,
		"size_gigabytes":   int(math.Round(float64(raid.Size) / GigaBytes)),
		"spare_devices":    spareDevices,
		"spare_partitions": sparePartitions,
	}

	if err := setTerraformState(d, tfstate); err != nil {
		return diag.Errorf("Could not set RAID state: %v", err)
	}

	return nil
}

func resourceRAIDUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Check the RAID configuration is valid
	if err = verifyRAIDConfig(client, machine, d); err != nil {
		return diag.FromErr(err)
	}

	// devices that are moving from active to spare or vice-versa need to be removed and then re-added in a separate call
	// we use the following function calls to determine the disks that have been moved vs. newly added/removed
	newBlockDevice, movedBlockDevice, removedBlockDevice := getMovedDevices(d, "block_devices", "spare_devices")
	newSpareDevice, movedSpareDevice, removedSpareDevice := getMovedDevices(d, "spare_devices", "block_devices")
	newPartition, movedPartition, removedPartition := getMovedDevices(d, "partitions", "spare_partitions")
	newSparePartition, movedSparePartition, removedSparePartition := getMovedDevices(d, "spare_partitions", "partitions")

	// We need to be very careful about order of operations, so that we maintain the minimum number of active disks required for the RAID level
	// We also need to ensure a disk is never included in both an add and a remove operation, as that will cause collisions in MAAS
	// This requires, at most, five separate update operations, in the worst case where active and spare are swapped:
	// add new disks, remove spare disks, add spare->active, remove active, add active->spare

	// 1. add all new active and spare disks
	raid, err := client.RAID.Update(machine.SystemID, id,
		&entity.RAIDUpdateParams{
			Name:               d.Get("name").(string),
			AddBlockDevices:    newBlockDevice,
			AddPartitions:      newPartition,
			AddSpareDevices:    newSpareDevice,
			AddSparePartitions: newSparePartition,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// 2. Remove any spare disks that are no longer in the RAID
	//    (Including disks moving from spare to active, so that we do not have add/remove collisions)
	if len(removedSpareDevice)+len(removedSparePartition) > 0 {
		if _, err = client.RAID.Update(machine.SystemID, id,
			&entity.RAIDUpdateParams{
				RemoveSpareDevices:    removedSpareDevice,
				RemoveSparePartitions: removedSparePartition,
			},
		); err != nil {
			return diag.FromErr(err)
		}
	}

	// 3. Re-add any disks that were moved from spare to active
	if len(movedBlockDevice)+len(movedPartition) > 0 {
		if _, err = client.RAID.Update(machine.SystemID, id,
			&entity.RAIDUpdateParams{
				AddBlockDevices: movedBlockDevice,
				AddPartitions:   movedPartition,
			},
		); err != nil {
			return diag.FromErr(err)
		}
	}

	// 4. Remove any active disks that are no longer in the RAID
	//    (Including disks moving from active to spare, so that we do not have add/remove collisions)
	if len(removedBlockDevice)+len(removedPartition) > 0 {
		if _, err = client.RAID.Update(machine.SystemID, id,
			&entity.RAIDUpdateParams{
				RemoveBlockDevices: removedBlockDevice,
				RemovePartitions:   removedPartition,
			},
		); err != nil {
			return diag.FromErr(err)
		}
	}

	// 5. Re-add any disks that were moved from active to spare
	if len(movedSpareDevice)+len(movedSparePartition) > 0 {
		if _, err = client.RAID.Update(machine.SystemID, id,
			&entity.RAIDUpdateParams{
				AddSpareDevices:    movedSpareDevice,
				AddSparePartitions: movedSparePartition,
			},
		); err != nil {
			return diag.FromErr(err)
		}
	}

	// We can finally perform mounting and formatting operations on the virtual block device
	if _, err = formatAndMountVirtualBlockDevice(client, &raid.VirtualDevice, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", raid.ID))

	return resourceRAIDRead(ctx, d, meta)
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

func verifyRAIDConfig(client *client.Client, machine *entity.Machine, d *schema.ResourceData) error {
	// Ensure the provided config has the correct disks for the RAID level, that each block device is partition-less,
	// that the boot-disk is not provided while block devices are (see: VolumeGroups for similar behavior), and that
	// provided disks are not included as both active and spare simultaneously.
	blockDevices := convertToStringSlice(d.Get("block_devices").(*schema.Set).List())
	spareDevices := convertToStringSlice(d.Get("spare_devices").(*schema.Set).List())
	partitions := convertToStringSlice(d.Get("partitions").(*schema.Set).List())
	sparePartitions := convertToStringSlice(d.Get("spare_partitions").(*schema.Set).List())

	// If any of the supplied block devices are the boot disk, MAAS will create partitions on all of the block devices
	// We ensure, similar to VolumeGroup, that the boot disk is not supplied as a block device, spare or active
	if err := verifyRAIDBootDevice(
		client,
		machine,
		append(blockDevices, spareDevices...),
		append(partitions, sparePartitions...),
	); err != nil {
		return err
	}

	// MAAS returns an unhelpful error if you supply a block device that has partitions, so
	// perform the check and turn it into a more helpful error that informs the user the changes they should make
	if err := verifyRAIDPartitionlessBlockDevices(
		client,
		machine,
		append(blockDevices, spareDevices...),
	); err != nil {
		return err
	}

	// verify the RAID Level is valid for the number of active and spare disks
	if err := verifyRAIDDevicesLevel(
		d.Get("level").(string),
		len(blockDevices)+len(partitions),
		len(spareDevices)+len(sparePartitions),
	); err != nil {
		return err
	}

	// verify disks are not included in both active and spare, as that causes MAAS issues
	if err := verifyRAIDcollision(
		blockDevices,
		spareDevices,
		partitions,
		sparePartitions,
	); err != nil {
		return err
	}

	return nil
}

func verifyRAIDBootDevice(client *client.Client, machine *entity.Machine, blockDevices []string, partitions []string) error {
	// If any of the block devices supplied to the RAID are the boot disk, MAAS will create
	// partitions on top of all of them. To prevent a Terraform error, we perform the same
	// check as in VolumeGroups, and ensure the boot disk is not a supplied block device.
	bootDisk := fmt.Sprintf("%v", machine.BootDisk.ID)

	var blockDeviceDisks []string

	for _, blockDevice := range machine.BlockDeviceSet {
		bdID := fmt.Sprintf("%d", blockDevice.ID)
		// add unique block devices in the RAID
		if slices.Contains(blockDevices, bdID) && !slices.Contains(blockDeviceDisks, bdID) {
			blockDeviceDisks = append(blockDeviceDisks, bdID)
			continue
		}
		// check if any partitions are present, and add the block device too
		for _, partition := range blockDevice.Partitions {
			partID := fmt.Sprintf("%d", partition.ID)
			if slices.Contains(partitions, partID) && !slices.Contains(blockDeviceDisks, bdID) {
				blockDeviceDisks = append(blockDeviceDisks, bdID)
				break
			}
		}
	}

	// If the boot disk is a part of the RAID, we need to ensure there are no block devices provided too
	if slices.Contains(blockDeviceDisks, bootDisk) && len(blockDevices) > 0 {
		return fmt.Errorf(
			"cannot construct a RAID with block devices if the boot disk %v (%v) is participating. Provide partitions on top of provided block devices instead",
			bootDisk,
			machine.BootDisk.Name,
		)
	}

	return nil
}

func verifyRAIDDevicesLevel(level string, activeCount int, spareCount int) error {
	// Ensure the number of provided active disks exeeds or matches what is required by the RAID level
	if activeCount <= 1 {
		return fmt.Errorf("RAIDs require at least two active disks")
	}

	if (level == "5" || level == "10") && activeCount < 3 {
		return fmt.Errorf("RAID level %v requires at least three active disks", level)
	}

	if level == "6" && activeCount < 4 {
		return fmt.Errorf("RAID level %v requires at least four active disks", level)
	}

	if level == "0" && spareCount > 0 {
		return fmt.Errorf("RAID level %v cannot use hot spares, supply active disks only", level)
	}

	// We won't stop the user, but we will warn them about atypical setups

	if level == "1" && spareCount > 1 {
		log.Printf("[WARN] RAID level %v with %d spares is unusual - only one spare is used during recovery\n", level, spareCount)
	}

	if level == "5" && spareCount > 1 {
		log.Printf("[WARN] RAID level %v with %d spares might not be the most fault tolerant topology - have you considered RAID 6 with %d spares instead?\n", level, spareCount, spareCount-1)
	}

	if spareCount > activeCount {
		log.Printf("[WARN] RAID has more spares (%d) than active disks (%d) - is this intentional?\n", spareCount, activeCount)
	}

	return nil
}

func verifyRAIDPartitionlessBlockDevices(client *client.Client, machine *entity.Machine, devices []string) error {
	// Ensure no block devices that have partitions have been supplied to the RAID
	for _, blockDevice := range machine.BlockDeviceSet {
		id := fmt.Sprintf("%d", blockDevice.ID)
		if slices.Contains(devices, id) && len(blockDevice.Partitions) > 0 {
			return fmt.Errorf("cannot create a RAID from a block device with partitions, supply the partitions for %v instead", blockDevice.Name)
		}
	}

	return nil
}

func verifyRAIDcollision(blockDevices []string, spareDevices []string, partitions []string, sparePartitions []string) error {
	// Ensure disks are not specified as both active and spare
	for _, disk := range blockDevices {
		if slices.Contains(spareDevices, disk) {
			return fmt.Errorf("cannot include block device %v as both active and spare, specify only a single location for the disk", disk)
		}
	}

	for _, disk := range partitions {
		if slices.Contains(sparePartitions, disk) {
			return fmt.Errorf("cannot include partition %v as both active and spare, specify only a single location for the disk", disk)
		}
	}

	return nil
}

func splitDeviceTypes(devices []entity.RAIDDevice) ([]string, []string, error) {
	// Split the disks into partitions and block devices by reading the device type
	var blockDevices []string

	var partitions []string

	for _, device := range devices {
		id := fmt.Sprintf("%d", device.ID)

		switch device.Type {
		case "physical":
			blockDevices = append(blockDevices, id)
		case "partition":
			partitions = append(partitions, id)
		default:
			return []string{}, []string{}, fmt.Errorf("device %v has an unknown type: %v", device.Name, device.Type)
		}
	}

	return blockDevices, partitions, nil
}

func getMovedDevices(d *schema.ResourceData, field string, counterpartField string) ([]string, []string, []string) {
	// Determine the list of disks, for a supplied field, that have been newly added, or newly removed. Additionally, determine
	// if any of the new disks are actually new, or have been moved from the counterpart field.
	// i.e.: the set of new active disks, removed active disks, or spare disks that have been moved to active.
	var createdDevices []string

	var movedDevices []string

	var removeDevices []string

	// if our field has no changes, evidently no devices are created, moved, or removed.
	if d.HasChange(field) {
		// determine the changes to our field
		oldDevices, newDevices := d.GetChange(field)
		oldList := convertToStringSlice(oldDevices.(*schema.Set).List())
		newList := convertToStringSlice(newDevices.(*schema.Set).List())

		// we also need a list of devices removed from the counterpart field to determine if
		// they were moved here
		counterpartOld, _ := d.GetChange(counterpartField)
		counterpartOldList := convertToStringSlice(counterpartOld.(*schema.Set).List())

		// if a new device exists in the old counterpart, it must have been moved
		// else, if it doesn't exist in the old list, it is newly created
		for _, device := range newList {
			if slices.Contains(counterpartOldList, device) {
				movedDevices = append(movedDevices, device)
			} else if !slices.Contains(oldList, device) {
				createdDevices = append(createdDevices, device)
			}
		}

		// if an old device doesnt exist in the new list, it must have been removed
		for _, device := range oldList {
			if !slices.Contains(newList, device) {
				removeDevices = append(removeDevices, device)
			}
		}
	}

	return createdDevices, movedDevices, removeDevices
}
