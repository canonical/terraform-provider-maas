package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASRAID_basic(t *testing.T) {
	var raid entity.RAID

	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := "raid_bd1"
	blockDevice2Name := "raid_bd2"
	blockDevice3Name := "raid_bd3"
	blockDevice4Name := "raid_bd4"

	// RAID 1 has the smallest disk requirement that still allows testing hot spares.
	level := "1"

	name := "test raid"
	fsType := "ext4"
	mountPoint := "/var/raidtest"

	changedName := "test renamed raid"
	changedFsType := "fat32"
	changedMountPoint := "/var/newraidtest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheclMAASRAIDDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: testAccRAIDMachine(machine) +
					// we include a seperate unused boot disk to avoid the boot disk/partition behaviour
					testAccRAIDBlockDevice("boot", true) +
					testAccRAIDBlockDevice(blockDevice1Name, false) +
					testAccRAIDPartition(blockDevice2Name, false) +
					testAccRAIDConfig(name, level, fsType, mountPoint,
						generateRAIDBlockDevices([]string{blockDevice1Name}),
						generateRAIDPartitions([]string{blockDevice2Name}),
						[]string{},
						[]string{},
					),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test", &raid),
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
				Config: testAccRAIDMachine(machine) +
					testAccRAIDBlockDevice("boot", true) +
					testAccRAIDBlockDevice(blockDevice1Name, false) +
					testAccRAIDPartition(blockDevice2Name, false) +
					testAccRAIDPartition(blockDevice3Name, false) +
					testAccRAIDBlockDevice(blockDevice4Name, false) +
					testAccRAIDConfig(changedName, level, changedFsType, changedMountPoint,
						generateRAIDBlockDevices([]string{blockDevice1Name}),
						generateRAIDPartitions([]string{blockDevice3Name}),
						generateRAIDBlockDevices([]string{blockDevice4Name}),
						generateRAIDPartitions([]string{blockDevice2Name}),
					),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test", &raid),
					resource.TestCheckResourceAttr("maas_raid.test", "name", changedName),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", changedFsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", changedMountPoint),

					resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "1"),

					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice3Name), "partitions.0.id"),
					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "spare_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice4Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "spare_partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "partitions.0.id"),
				),
			},
			// Test the worst-case operation to ensure update is working correctly: fully swapping active and spare disks
			{
				Config: testAccRAIDMachine(machine) +
					testAccRAIDBlockDevice("boot", true) +
					testAccRAIDBlockDevice(blockDevice1Name, false) +
					testAccRAIDPartition(blockDevice2Name, false) +
					testAccRAIDPartition(blockDevice3Name, false) +
					testAccRAIDBlockDevice(blockDevice4Name, false) +
					testAccRAIDConfig(changedName, level, changedFsType, changedMountPoint,
						generateRAIDBlockDevices([]string{blockDevice4Name}),
						generateRAIDPartitions([]string{blockDevice2Name}),
						generateRAIDBlockDevices([]string{blockDevice1Name}),
						generateRAIDPartitions([]string{blockDevice3Name}),
					),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test", &raid),
					resource.TestCheckResourceAttr("maas_raid.test", "name", changedName),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", changedFsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", changedMountPoint),

					resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "1"),

					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice4Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "partitions.0.id"),
					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "spare_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "spare_partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice3Name), "partitions.0.id"),
				),
			},
		},
	})
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

func testAccRAIDBlockDevice(name string, isBoot bool) string {
	return fmt.Sprintf(`
resource "maas_block_device" "%v" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 2
  block_size     = 512
  id_path        = "/dev/%v"
  is_boot_device = %t
}
`, name, name, name, isBoot)
}

func testAccRAIDPartition(name string, isBoot bool) string {
	return fmt.Sprintf(`
resource "maas_block_device" "%v" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 3
  block_size     = 512
  id_path        = "/dev/%v"
  is_boot_device = %t

  partitions {
    size_gigabytes = 2
  }
}
`, name, name, name, isBoot)
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

func testAccCheckRAIDExists(rn string, raid *entity.RAID) resource.TestCheckFunc {
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

		gotRAID, err := conn.RAID.Get(machine, id)
		if err != nil {
			return fmt.Errorf("error getting the RAID: %s", err)
		}

		*raid = *gotRAID

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
