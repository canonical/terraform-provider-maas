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

func TestAccResourceMaasVolumeGroup_basic(t *testing.T) {
	var volumeGroup entity.VolumeGroup

	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	name := "test volume group"

	checks := []resource.TestCheckFunc{
		testAccCheckMaasVolumeGroupExists("maas_volume_group.test", &volumeGroup),
		resource.TestCheckResourceAttr("maas_volume_group.test", "name", name),
		resource.TestCheckResourceAttr("maas_volume_group.test", "block_devices.#", "2"),
		resource.TestCheckResourceAttrPair("maas_volume_group.test", "block_devices.0", "maas_block_device.bd1", "id"),
		resource.TestCheckResourceAttrPair("maas_volume_group.test", "block_devices.1", "maas_block_device.bd2", "id"),
		resource.TestCheckResourceAttrPair("maas_volume_group.test", "machine", "data.maas_machine.machine", "id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasVolumeGroupDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasVolumeGroup(machine, name),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMaasVolumeGroup(machine string, name string) string {
	return fmt.Sprintf(`

data "maas_machine" "machine" {
  hostname = "%s"
}

resource "maas_block_device" "bd1" {
  machine        = data.maas_machine.machine.id
  name           = "bd1"
  size_gigabytes = 20
  block_size     = 512
  id_path        = "/dev/bd1"
}

resource "maas_block_device" "bd2" {
  machine        = data.maas_machine.machine.id
  name           = "bd2"
  size_gigabytes = 50
  block_size     = 512
  id_path        = "/dev/bd2"
}

resource "maas_volume_group" "test" {
 	machine       = data.maas_machine.machine.id
	name          = "%s"
	block_devices = [maas_block_device.bd1.id, maas_block_device.bd2.id]

	depends_on = [maas_block_device.bd1, maas_block_device.bd2]
}

`, machine, name)
}

func testAccCheckMaasVolumeGroupExists(rn string, volumeGroup *entity.VolumeGroup) resource.TestCheckFunc {
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

		gotVolumeGroup, err := conn.VolumeGroup.Get(machine, id)
		if err != nil {
			return fmt.Errorf("error getting volume group: %s", err)
		}

		*volumeGroup = *gotVolumeGroup

		return nil
	}
}

func testAccCheckMaasVolumeGroupDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_volume_group" {
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

		response, err := conn.VolumeGroup.Get(machine, id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("Volume group %s (%d) still exists.", response.Name, id)
			}
		}

		// 404 means destroyed, anything else is an error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}
	return nil
}
