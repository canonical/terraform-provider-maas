package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASRAID_basic(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := "raid_bd1"
	blockDevice2Name := "raid_bd2"
	blockDevice3Name := "raid_bd3"
	blockDevice4Name := "raid_bd4"

	// RAID 1 has the smallest disk requirement that still allows testing hot spares.
	level := "1"

	name := "test RAID"
	fsType := "ext4"
	mountPoint := "/var/raidtest"

	changedName := "test RAID renamed"
	changedFsType := "fat32"
	changedMountPoint := "/var/raidrename"

	swappedName := "test RAID swapped active/spare"
	swappedFsType := "ext4"
	swappedMountPoint := "/var/raidswap"

	// we include a seperate unused boot disk to avoid the boot disk/partition behaviour
	baseConfig := testAccRAIDMachine(machine) +
		testAccRAIDBlockDevice("boot", 2, true) +
		testAccRAIDBlockDevice(blockDevice1Name, 2, false) +
		testAccRAIDPartition(blockDevice2Name, 2, false) +
		testAccRAIDBlockDevice(blockDevice3Name, 2, false) +
		testAccRAIDPartition(blockDevice4Name, 2, false)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheclMAASRAIDDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: baseConfig + testAccRAIDConfig(name, level, fsType, mountPoint,
					generateRAIDBlockDevices([]string{blockDevice1Name}),
					generateRAIDPartitions([]string{blockDevice2Name}),
					[]string{},
					[]string{},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test"),
					resource.TestCheckResourceAttr("maas_raid.test", "name", name),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", fsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", mountPoint),

					resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "0"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "0"),

					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "partitions.0.id"),
				),
			},
			// Test updating the fields, moving an active disk, and adding a spare
			{
				Config: baseConfig +
					testAccRAIDConfig(changedName, level, changedFsType, changedMountPoint,
						generateRAIDBlockDevices([]string{blockDevice1Name}),
						generateRAIDPartitions([]string{blockDevice4Name}),
						generateRAIDBlockDevices([]string{blockDevice3Name}),
						generateRAIDPartitions([]string{blockDevice2Name}),
					),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test"),
					resource.TestCheckResourceAttr("maas_raid.test", "name", changedName),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", changedFsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", changedMountPoint),

					resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "1"),

					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice4Name), "partitions.0.id"),
					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "spare_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice3Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "spare_partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "partitions.0.id"),
				),
			},
			// Test the worst-case operation to ensure update is working correctly: fully swapping active and spare disks
			{
				Config: baseConfig +
					testAccRAIDConfig(swappedName, level, swappedFsType, swappedMountPoint,
						generateRAIDBlockDevices([]string{blockDevice3Name}),
						generateRAIDPartitions([]string{blockDevice2Name}),
						generateRAIDBlockDevices([]string{blockDevice1Name}),
						generateRAIDPartitions([]string{blockDevice4Name}),
					),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test"),
					resource.TestCheckResourceAttr("maas_raid.test", "name", swappedName),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", swappedFsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", swappedMountPoint),

					resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "1"),

					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice3Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "partitions.0.id"),
					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "spare_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "spare_partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice4Name), "partitions.0.id"),
				),
			},
		},
	})
}

func TestAccResourceMAASRAID_differentLevels(t *testing.T) {
	// We need to test a RAID can be created for every level supported
	// TODO: Update this when LP:2109708 is released in MAAS
	validRAIDLevels := []string{"0", "1", "5", "6"}

	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")

	fsType := "ext4"

	for _, testLevel := range validRAIDLevels {
		thisLevel := testLevel // capture range variable
		t.Run(fmt.Sprintf("RAID_level_%s", thisLevel), func(t *testing.T) {
			thisName := fmt.Sprintf("test RAID level %s", thisLevel)
			thisMount := fmt.Sprintf("/var/test_raid_%s", thisLevel)
			blockDevice1Name := fmt.Sprintf("raid_level_%s_test_bd1", thisLevel)
			blockDevice2Name := fmt.Sprintf("raid_level_%s_test_bd2", thisLevel)
			blockDevice3Name := fmt.Sprintf("raid_level_%s_test_bd3", thisLevel)
			blockDevice4Name := fmt.Sprintf("raid_level_%s_test_bd4", thisLevel)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
				Providers:    testutils.TestAccProviders,
				CheckDestroy: testAccCheclMAASRAIDDestroy,
				ErrorCheck:   func(err error) error { return err },
				Steps: []resource.TestStep{
					{
						Config: testAccRAIDMachine(machine) +
							testAccRAIDBlockDevice(fmt.Sprintf("boot_device_%s", thisLevel), 2, true) +
							testAccRAIDBlockDevice(blockDevice1Name, 2, false) +
							testAccRAIDBlockDevice(blockDevice2Name, 2, false) +
							testAccRAIDPartition(blockDevice3Name, 2, false) +
							testAccRAIDPartition(blockDevice4Name, 2, false) +
							testAccRAIDConfig(thisName, thisLevel, fsType, thisMount,
								generateRAIDBlockDevices([]string{blockDevice1Name, blockDevice2Name}),
								generateRAIDPartitions([]string{blockDevice3Name, blockDevice4Name}),
								[]string{},
								[]string{},
							),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckRAIDExists("maas_raid.test"),
							resource.TestCheckResourceAttr("maas_raid.test", "name", thisName),
							resource.TestCheckResourceAttr("maas_raid.test", "level", thisLevel),
							resource.TestCheckResourceAttr("maas_raid.test", "fs_type", fsType),
							resource.TestCheckResourceAttr("maas_raid.test", "mount_point", thisMount),
							resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "2"),
							resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "2"),
							resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "0"),
							resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "0"),
							resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.*", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
							resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.*", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "id"),
							resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "partitions.*", fmt.Sprintf("maas_block_device.%v", blockDevice3Name), "partitions.0.id"),
							resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "partitions.*", fmt.Sprintf("maas_block_device.%v", blockDevice4Name), "partitions.0.id"),
						),
					},
				},
			})
		})
	}
}

func generateRAIDBlockDevices(devices []string) []string {
	var output []string
	for _, device := range devices {
		output = append(output, fmt.Sprintf("maas_block_device.%v.id", device))
	}

	return output
}
func generateRAIDPartitions(partitions []string) []string {
	var output []string
	for _, part := range partitions {
		output = append(output, fmt.Sprintf("maas_block_device.%v.partitions.0.id", part))
	}

	return output
}

func testAccRAIDMachine(machine string) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = %q
}
`, machine)
}

func testAccRAIDBlockDevice(name string, size int, isBoot bool) string {
	return fmt.Sprintf(`
resource "maas_block_device" "%v" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = %d
  block_size     = 512
  id_path        = "/dev/%v"
  is_boot_device = %t
}
`, name, name, size, name, isBoot)
}

func testAccRAIDPartition(name string, size int, isBoot bool) string {
	return fmt.Sprintf(`
resource "maas_block_device" "%v" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = %d
  block_size     = 512
  id_path        = "/dev/%v"
  is_boot_device = %t

  partitions {
    size_gigabytes = %d
  }
}
`, name, name, size+1, name, isBoot, size)
}

func sliceToString(devices []string) string {
	device := fmt.Sprintf("[%s]", strings.Join(func() []string {
		s := make([]string, len(devices))
		for i, v := range devices {
			s[i] = fmt.Sprintf("%v", v)
		}

		return s
	}(), ", "))

	return device
}

func testAccRAIDConfig(name string, level string, fsType string, mountPoint string, blockDevices []string, partitions []string, spareDevices []string, sparePartitions []string) string {
	return fmt.Sprintf(`
resource "maas_raid" "test" {
  machine     = data.maas_machine.machine.id
  name	      = %q
  level       = %q
  fs_type     = %q
  mount_point = %q

  block_devices    = %v
  partitions       = %v
  spare_devices    = %v
  spare_partitions = %v
}
`, name, level, fsType, mountPoint, sliceToString(blockDevices), sliceToString(partitions), sliceToString(spareDevices), sliceToString(sparePartitions))
}

func testAccCheckRAIDExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		machine, ok := rs.Primary.Attributes["machine"]
		if !ok {
			return fmt.Errorf("Could not find machine id on resource")
		}

		if _, err = conn.RAID.Get(machine, id); err != nil {
			return fmt.Errorf("error getting the RAID: %s", err)
		}

		return nil
	}
}

func testAccCheclMAASRAIDDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_raid" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		machine, ok := rs.Primary.Attributes["machine"]
		if !ok {
			return fmt.Errorf("Could not find machine id on resource")
		}

		response, err := conn.RAID.Get(machine, id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("RAID %s (%d) still exists.", response.Name, id)
			}
		}

		// 404 means destroyed, anything else is an error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
