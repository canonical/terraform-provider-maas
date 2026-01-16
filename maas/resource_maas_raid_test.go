package maas_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASRAID_basic(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := acctest.RandomWithPrefix("tf-raid-bd")
	blockDevice2Name := acctest.RandomWithPrefix("tf-raid-bd")
	blockDevice3Name := acctest.RandomWithPrefix("tf-raid-bd")
	blockDevice4Name := acctest.RandomWithPrefix("tf-raid-bd")

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

	// we include a separate unused boot disk to avoid the boot disk/partition behavior
	baseConfig := testAccRAIDMachine(machine) +
		testAccRAIDBlockDevice(acctest.RandomWithPrefix("boot"), 2, true) +
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

func TestAccResourceMAASRAID_formatAndMount(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := acctest.RandomWithPrefix("tf-raid-bd")
	blockDevice2Name := acctest.RandomWithPrefix("tf-raid-bd")
	blockDevice3Name := acctest.RandomWithPrefix("tf-raid-bd")
	blockDevice4Name := acctest.RandomWithPrefix("tf-raid-bd")

	// RAID 1 has the smallest disk requirement that still allows testing hot spares.
	level := "1"

	// Test 1: `fs_type` not specified
	test1FsType := ""
	test1MountPoint := "/var/raidtest"

	// Test 2: `mount_point` not specified
	test2FsType := "ext4"
	test2MountPoint := ""

	// we include a separate unused boot disk to avoid the boot disk/partition behavior
	baseConfig := testAccRAIDMachine(machine) +
		testAccRAIDBlockDevice(acctest.RandomWithPrefix("boot"), 2, true) +
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
				Config: baseConfig + testAccRAIDConfig("RAID", level, test1FsType, test1MountPoint,
					generateRAIDBlockDevices([]string{blockDevice1Name}),
					generateRAIDPartitions([]string{blockDevice2Name}),
					[]string{},
					[]string{},
				),
				ExpectError: regexp.MustCompile(`invalid block device mount configuration: fs_type must be specified when mount_point is set`),
			},
			{
				Config: baseConfig + testAccRAIDConfig("RAID", level, test2FsType, test2MountPoint,
					generateRAIDBlockDevices([]string{blockDevice1Name}),
					generateRAIDPartitions([]string{blockDevice2Name}),
					[]string{},
					[]string{},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test"),
					resource.TestCheckResourceAttr("maas_raid.test", "name", "RAID"),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", test2FsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", test2MountPoint),

					resource.TestCheckResourceAttr("maas_raid.test", "block_devices.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "partitions.#", "1"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_devices.#", "0"),
					resource.TestCheckResourceAttr("maas_raid.test", "spare_partitions.#", "0"),

					resource.TestCheckTypeSetElemAttrPair("maas_raid.test", "block_devices.0", fmt.Sprintf("maas_block_device.%v", blockDevice1Name), "id"),
					resource.TestCheckResourceAttrPair("maas_raid.test", "partitions.0", fmt.Sprintf("maas_block_device.%v", blockDevice2Name), "partitions.0.id"),
				),
			},
		},
	})
}

func TestAccResourceMAASRAID_differentLevels(t *testing.T) {
	validRAIDLevels := []string{"0", "1", "5", "6", "10"}

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

func TestVerifyRAIDDevicesLevel(t *testing.T) {
	// define the test cases
	testMatrix := []struct {
		name         string
		level        string
		activeCount  int
		spareCount   int
		hasError     bool
		errorMessage string
		hasLog       bool
		logMessage   string
	}{
		// Test minimum disk requirement
		{"RAID too few disks", "1", 1, 0, true, "require at least two active disks", false, ""},

		// Test each RAID level
		{"RAID 0 valid", "0", 2, 0, false, "", false, ""},
		{"RAID 0 with spare", "0", 2, 1, true, "cannot use hot spares", false, ""},
		{"RAID 0 too few disks", "0", 1, 0, true, "require at least two active disks", false, ""},

		{"RAID 1 valid", "1", 2, 0, false, "", false, ""},
		{"RAID 1 valid spares", "1", 2, 1, false, "", false, ""},
		{"RAID 1 too few disks", "1", 1, 0, true, "require at least two active disks", false, ""},

		{"RAID 5 valid", "5", 3, 0, false, "", false, ""},
		{"RAID 5 valid spares", "5", 3, 1, false, "", false, ""},
		{"RAID 5 too few disks", "5", 2, 0, true, "requires at least three active disks", false, ""},

		{"RAID 6 valid", "6", 4, 0, false, "", false, ""},
		{"RAID 6 valid spares", "6", 4, 4, false, "", false, ""},
		{"RAID 6 too few disks", "6", 3, 0, true, "requires at least four active disks", false, ""},

		{"RAID 10 valid", "10", 3, 0, false, "", false, ""},
		{"RAID 10 valid spares", "10", 3, 2, false, "", false, ""},
		{"RAID 10 too few disks", "10", 2, 0, true, "requires at least three active disks", false, ""},

		// These shouldn't produce an error, only a usage warning
		{"RAID 1 unusual spares", "1", 10, 4, false, "", true, "spares is unusual"},
		{"RAID 5 valid spares", "5", 10, 4, false, "", true, "have you considered RAID 6"},
		{"RAID more spares than active", "6", 4, 10, false, "", true, "more spares (10) than active disks"},
	}

	for _, thisTest := range testMatrix {
		t.Run(thisTest.name, func(t *testing.T) {
			// we want to capture log outputs to test warnings
			var logBuffer bytes.Buffer

			log.SetOutput(&logBuffer)

			defer log.SetOutput(os.Stderr)

			err := verifyRAIDDevicesLevel(thisTest.level, thisTest.activeCount, thisTest.spareCount)
			if thisTest.hasError {
				if err == nil {
					t.Errorf("expected error %q, but function passed", thisTest.errorMessage)
				} else if !strings.Contains(err.Error(), thisTest.errorMessage) {
					t.Errorf("expected error %q, but got %q", thisTest.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			logOutput := logBuffer.String()

			if thisTest.hasLog {
				if logOutput == "" {
					t.Errorf("expected log message %q, but returned nothing", thisTest.logMessage)
				} else if !strings.Contains(logOutput, thisTest.logMessage) {
					t.Errorf("expected log message %q, but got %q", thisTest.logMessage, logBuffer.String())
				}
			} else if logOutput != "" {
				t.Errorf("unexpected log message: %q", logOutput)
			}
		})
	}
}

func TestVerifyRAIDCollisions(t *testing.T) {
	// define the test cases
	testMatrix := []struct {
		name            string
		blockDevices    []string
		spareDevices    []string
		partitions      []string
		sparePartitions []string
		hasError        bool
		errorMessage    string
	}{
		{"no disks", []string{}, []string{}, []string{}, []string{}, false, ""},
		{"no collisions", []string{"bd1"}, []string{"bd2"}, []string{"p1"}, []string{"p2"}, false, ""},
		// minimum collision case
		{"block device collision", []string{"bd1"}, []string{"bd1"}, []string{}, []string{}, true, "cannot include block device bd1 as both active and spare"},
		{"partition collision", []string{}, []string{}, []string{"p1"}, []string{"p1"}, true, "cannot include partition p1 as both active and spare"},
		// all unique, multiple inputs
		{"multiple block devices", []string{"bd1", "bd2"}, []string{"bd3", "bd4"}, []string{"p1"}, []string{"p2"}, false, ""},
		{"multiple partitions", []string{"bd1"}, []string{"bd2"}, []string{"p1", "p2"}, []string{"p3", "p4"}, false, ""},
		// two valid entries, one collision
		{"overlapping block device", []string{"bd1", "bd2"}, []string{"bd2", "bd3"}, []string{}, []string{}, true, "cannot include block device bd2 as both active and spare"},
		{"overlapping partition", []string{}, []string{}, []string{"p1", "p2"}, []string{"p2", "p3"}, true, "cannot include partition p2 as both active and spare"},
		// The block device check occurs first, so will trigger before checking partitions
		{"block device and partition collision", []string{"bd1"}, []string{"bd1"}, []string{"p1"}, []string{"p1"}, true, "cannot include block device bd1 as both active and spare"},
	}

	for _, thisTest := range testMatrix {
		t.Run(thisTest.name, func(t *testing.T) {
			err := verifyRAIDcollision(thisTest.blockDevices, thisTest.spareDevices, thisTest.partitions, thisTest.sparePartitions)
			if thisTest.hasError {
				if err == nil {
					t.Errorf("expected error %q, but function passed", thisTest.errorMessage)
				} else if !strings.Contains(err.Error(), thisTest.errorMessage) {
					t.Errorf("expected error %q, but got %q", thisTest.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
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

// TODO: Determine a way of importing these functions from the maas_raid file directly,
// as duplication can lead to things becoming out of step

func verifyRAIDDevicesLevel(level string, activeCount int, spareCount int) error {
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

func verifyRAIDcollision(blockDevices []string, spareDevices []string, partitions []string, sparePartitions []string) error {
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
