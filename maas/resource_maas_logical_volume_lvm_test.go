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

func TestAccResourceMAASLogicalVolumeLvm_basic(t *testing.T) {
	var LogicalVolumeLVM entity.BlockDevice

	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")

	fsType := "ext4"
	name := "LVM test"
	size := 69

	checks := []resource.TestCheckFunc{
		testAccCheckMAASLogicalVolumeLvmExists("maas_logical_volume_lvm.test", &LogicalVolumeLVM),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASLogicalVolumeDestroy,
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: testAccMAASLogicalVolumeLvm(machine, fsType, name, size),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASLogicalVolumeLvm(machine string, fsType string, name string, size int) string {
	return fmt.Sprintf(`
%s

resource "maas_logical_volume_lvm" "test" {
  fs_type 		 = %q
  machine 		 = data.maas_machine.machine.id
  name 			 = %q
  size_gigabytes = %d
  volume_group 	 = maas_volume_group.test.id
}
`, testAccMAASVolumeGroup(machine, "test-vg", []string{"maas_block_device.bd2.id"}, []string{"maas_block_device.bd1.partitions.0.id"}),
		fsType, name, size)
}

func testAccCheckMAASLogicalVolumeLvmExists(rn string, logicalVolume *entity.BlockDevice) resource.TestCheckFunc {
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
