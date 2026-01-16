package maas_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASLogicalVolume_basic(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := acctest.RandomWithPrefix("tf")
	blockDevice2Name := acctest.RandomWithPrefix("tf")
	volumeGroupName := acctest.RandomWithPrefix("tf")

	fsType := "ext4"
	name := "LVM test"
	mountPoint := "/var/test"
	size := 5

	changedFsType := "fat32"
	changedName := "LVM updated"
	changedMountPoint := "/var/changed"
	changedSize := 2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckLogicalVolumeDestroy,
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: testAccLogicalVolume(blockDevice1Name, blockDevice2Name, volumeGroupName, machine, fsType, name, size, mountPoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicalVolumeExists("maas_logical_volume.test"),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "name", name),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "fs_type", fsType),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "mount_point", mountPoint),
				),
			},
			// Test the update function
			{
				Config: testAccLogicalVolume(blockDevice1Name, blockDevice2Name, volumeGroupName, machine, changedFsType, changedName, changedSize, changedMountPoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicalVolumeExists("maas_logical_volume.test"),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "name", changedName),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "fs_type", changedFsType),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "mount_point", changedMountPoint),
				),
			},
		},
	})
}

func TestAccResourceMAASLogicalVolume_formatAndMount(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := acctest.RandomWithPrefix("tf-lv-bd")
	blockDevice2Name := acctest.RandomWithPrefix("tf-lv-bd")
	volumeGroupName := acctest.RandomWithPrefix("tf-lv-vg")

	// Test 1: `fs_type` not specified
	test1FsType := ""
	test1MountPoint := "/var/test"

	// Test 2: `mount_point` not specified
	test2FsType := "fat32"
	test2MountPoint := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckLogicalVolumeDestroy,
		Steps: []resource.TestStep{
			// Test 1: `fs_type` not specified
			{
				Config:      testAccLogicalVolume(blockDevice1Name, blockDevice2Name, volumeGroupName, machine, test1FsType, "LVM test", 5, test1MountPoint),
				ExpectError: regexp.MustCompile(`invalid block device mount configuration: fs_type must be specified when mount_point is set`),
			},
			// Test 2: `mount_point` not specified
			{
				Config: testAccLogicalVolume(blockDevice1Name, blockDevice2Name, volumeGroupName, machine, test2FsType, "LVM test", 5, test2MountPoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicalVolumeExists("maas_logical_volume.test"),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "name", "LVM test"),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "fs_type", test2FsType),
					resource.TestCheckResourceAttr("maas_logical_volume.test", "mount_point", test2MountPoint),
				),
			},
		},
	})
}

func testAccLogicalVolume(bd1Name string, bd2Name string, vgName string, machine string, fsType string, name string, size int, mountPoint string) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = %q
}

resource "maas_block_device" "lvm_bd1" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 6
  block_size     = 512
  id_path        = "/dev/lvm_bd1"
  is_boot_device = true

  partitions {
    size_gigabytes = 5
  }
}

resource "maas_block_device" "lvm_bd2" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 5
  block_size     = 512
  id_path        = "/dev/bd2"
}

resource "maas_volume_group" "lvm_vg" {
  machine       = data.maas_machine.machine.id
  name          = %q
  block_devices = [maas_block_device.lvm_bd2.id]
  partitions    = [maas_block_device.lvm_bd1.partitions.0.id]
}

resource "maas_logical_volume" "test" {
  fs_type        = %q
  machine        = data.maas_machine.machine.id
  name           = %q
  volume_group   = maas_volume_group.lvm_vg.id
  size_gigabytes = %d
  mount_point    = %q
}
`, machine, bd1Name, bd2Name, vgName, fsType, name, size, mountPoint)
}

func testAccCheckLogicalVolumeExists(rn string) resource.TestCheckFunc {
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

		_, err = conn.BlockDevice.Get(machine, id)
		if err != nil {
			return fmt.Errorf("error getting the logical volume: %s", err)
		}

		return nil
	}
}

func testAccCheckLogicalVolumeDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_logical_volume" {
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

		response, err := conn.BlockDevice.Get(machine, id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("Logical Volume %s (%d) still exists.", response.Name, id)
			}
		}

		// 404 means destroyed, anything else is an error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
