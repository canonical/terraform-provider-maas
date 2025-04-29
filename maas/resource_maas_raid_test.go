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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASRAID_basic(t *testing.T) {
	var raid entity.RAID

	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")
	blockDevice1Name := acctest.RandomWithPrefix("tf")
	blockDevice2Name := acctest.RandomWithPrefix("tf")

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
				Config: testAccRAID(machine, blockDevice1Name, blockDevice2Name, name, level, fsType, mountPoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test", &raid),
					resource.TestCheckResourceAttr("maas_raid.test", "name", name),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", fsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", mountPoint),
				),
			},
			// Test basic update
			{
				Config: testAccRAID(machine, blockDevice1Name, blockDevice2Name, changedName, level, changedFsType, changedMountPoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRAIDExists("maas_raid.test", &raid),
					resource.TestCheckResourceAttr("maas_raid.test", "name", changedName),
					resource.TestCheckResourceAttr("maas_raid.test", "level", level),
					resource.TestCheckResourceAttr("maas_raid.test", "fs_type", changedFsType),
					resource.TestCheckResourceAttr("maas_raid.test", "mount_point", changedMountPoint),
				),
			},
			// TODO: Test updating block devices, partitions, and spares
		},
	})
}

func testAccRAID(machine string, bd1Name string, bd2Name string, name string, level string, fsType string, mountPoint string) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = %q
}

resource "maas_block_device" "raid_bd1" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 6
  block_size     = 512
  id_path        = "/dev/raid_bd1"
}

resource "maas_block_device" "raid_bd2" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 6
  block_size     = 512
  id_path        = "/dev/raid_bd2"

  partitions {
    size_gigabytes = 5
  }
}

resource "maas_raid" "test" {
  machine     = data.maas_machine.machine.id
  name	      = %q
  level       = %q
  fs_type     = %q
  mount_point = %q

  block_devices = [
  	maas_block_device.raid_bd1.id,
  ]
  partitions = [
	maas_block_device.raid_bd2.partitions.0.id
  ]
  
}
`, machine, bd1Name, bd2Name, name, level, fsType, mountPoint)
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
