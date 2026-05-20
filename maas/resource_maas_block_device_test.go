package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccMAASBlockDevice(machine string) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = "%s"
}

resource "maas_block_device" "test" {
  machine        = data.maas_machine.machine.id
  name           = "sda"
  size_gigabytes = 100
  block_size     = 512
  is_boot_device = true
  id_path        = "/dev/sda"

  tags = [
    "rotary", "1pm",
  ]

  partitions {
    size_gigabytes = 10
    fs_type        = "ext4"
    label          = "media"
    mount_point    = "/"
  }

  partitions {
    size_gigabytes = 11
    fs_type        = "ext4"
    mount_point    = "/var"
  }

  partitions {
    size_gigabytes = 12
    fs_type        = "ext4"
    mount_point    = "/var/log"
  }

  partitions {
    size_gigabytes = 13
    fs_type        = "ext4"
    mount_point    = "/var/adm"
  }

  partitions {
    size_gigabytes = 14
    fs_type        = "ext4"
    mount_point    = "/var/tmp"
  }

  partitions {
    size_gigabytes = 15
    fs_type        = "ext4"
    mount_point    = "/var/log/audit"
  }

  partitions {
    size_gigabytes = 16
    fs_type        = "ext4"
    mount_point    = "/tmp"
  }

}
`, machine)
}

func testAccMAASBlockDeviceWithIDPath(machine, name, idPath string) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = "%s"
}

resource "maas_block_device" "test" {
  machine        = data.maas_machine.machine.id
  name           = "%s"
  size_gigabytes = 5
  block_size     = 512
  id_path        = "%s"
}
`, machine, name, idPath)
}

func TestAccResourceMAASBlockDevice_basic(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("maas_block_device.test", "name", "sda"),
		resource.TestCheckResourceAttr("maas_block_device.test", "size_gigabytes", "100"),
		resource.TestCheckResourceAttr("maas_block_device.test", "block_size", "512"),
		resource.TestCheckResourceAttr("maas_block_device.test", "is_boot_device", "true"),
		resource.TestCheckResourceAttr("maas_block_device.test", "id_path", "/dev/sda"),
		resource.TestCheckResourceAttr("maas_block_device.test", "tags.#", "2"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.#", "7"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.0.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.0.mount_point", "/"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.0.size_gigabytes", "10"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.1.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.1.mount_point", "/var"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.1.size_gigabytes", "11"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.2.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.2.mount_point", "/var/log"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.2.size_gigabytes", "12"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.3.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.3.mount_point", "/var/adm"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.3.size_gigabytes", "13"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.4.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.4.mount_point", "/var/tmp"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.4.size_gigabytes", "14"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.5.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.5.mount_point", "/var/log/audit"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.5.size_gigabytes", "15"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.6.fs_type", "ext4"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.6.mount_point", "/tmp"),
		resource.TestCheckResourceAttr("maas_block_device.test", "partitions.6.size_gigabytes", "16"),
		resource.TestCheckResourceAttrPair("maas_block_device.test", "machine", "data.maas_machine.machine", "id"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASBlockDevice(machine),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccResourceMAASBlockDevice_stale(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")

	oldName := acctest.RandomWithPrefix("tf-disk")
	newName := acctest.RandomWithPrefix("tf-disk")
	idPath := acctest.RandomWithPrefix("/dev/vd")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASBlockDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMAASBlockDeviceWithIDPath(machine, oldName, idPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("maas_block_device.test", "name", oldName),
					resource.TestCheckResourceAttr("maas_block_device.test", "id_path", idPath),
				),
			},
			// rename the disk to check there is nothing stale remaining
			{
				Config: testAccMAASBlockDeviceWithIDPath(machine, newName, idPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("maas_block_device.test", "name", newName),
					resource.TestCheckResourceAttr("maas_block_device.test", "id_path", idPath),

					testAccCheckMAASBlockDeviceStaleRemoved("maas_block_device.test", oldName),
				),
			},
		},
	})
}

func testAccCheckMAASBlockDeviceStaleRemoved(rn string, oldName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		machineID := rs.Primary.Attributes["machine"]
		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		devices, err := client.BlockDevices.Get(machineID)
		if err != nil {
			return fmt.Errorf("error fetching block devices: %s", err)
		}

		for _, d := range devices {
			if d.Name == oldName {
				return fmt.Errorf("Stale Block device %q still exists", oldName)
			}
		}

		return nil
	}
}

func testAccCheckMAASBlockDeviceDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_block_device" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		machine := rs.Primary.Attributes["machine"]

		response, err := conn.BlockDevice.Get(machine, id)
		if err == nil {
			if response != nil {
				_ = conn.BlockDevice.Delete(machine, id)
			}

			continue
		}

		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
