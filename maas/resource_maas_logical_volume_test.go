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

func TestAccResourceMAASLogicalVolume_basic(t *testing.T) {
	var LogicalVolume entity.BlockDevice

	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")

	fsType := "ext4"
	name := "LVM test"
	size := 69

	checks := []resource.TestCheckFunc{
		testAccCheckMAASLogicalVolumeExists("maas_logical_volume_lvm.test", &LogicalVolume),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASLogicalVolumeDestroy,
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: testAccMAASLogicalVolume(machine, fsType, name, size),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASLogicalVolume(machine string, fsType string, name string, size int) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = %q
}

resource "maas_block_device" "lvm_bd1" {
  machine        = data.maas_machine.machine.id
  name           = "lvm_bd1"
  size_gigabytes = 25
  block_size     = 512
  id_path        = "/dev/lvm_bd1"
  is_boot_device = true

  partitions {
    size_gigabytes = 20
  }
}

resource "maas_block_device" "lvm_bd2" {
  machine        = data.maas_machine.machine.id
  name           = "bd2"
  size_gigabytes = 50
  block_size     = 512
  id_path        = "/dev/bd2"
}

resource "maas_volume_group" "lvm_vg" {
  machine       = data.maas_machine.machine.id
  name          = "test-vg"
  block_devices = [maas_block_device.lvm_bd2.id]
  partitions 	= [maas_block_device.lvm_bd1.partitions.0.id]
}

resource "maas_logical_volume_lvm" "test" {
  fs_type 		 = %q
  machine 		 = data.maas_machine.machine.id
  name 			 = %q
  size_gigabytes = %d
  volume_group 	 = maas_volume_group.lvm_vg.id
}
`, machine, fsType, name, size)
}

func testAccCheckMAASLogicalVolumeExists(rn string, logicalVolume *entity.BlockDevice) resource.TestCheckFunc {
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

		gotLogicalVolume, err := conn.BlockDevice.Get(machine, id)
		if err != nil {
			return fmt.Errorf("error getting the logical volume: %s", err)
		}

		*logicalVolume = *gotLogicalVolume

		return nil
	}
}

func testAccCheckMAASLogicalVolumeDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_logical_volume_lvm" {
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
				return fmt.Errorf("Logical Volumep %s (%d) still exists.", response.Name, id)
			}
		}

		// 404 means destroyed, anything else is an error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
