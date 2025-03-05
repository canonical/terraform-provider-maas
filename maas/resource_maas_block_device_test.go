package maas_test

import (
	"fmt"
	"os"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccMaasBlockDevice(machine string) string {
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

func TestAccResourceMaasBlockDevice_basic(t *testing.T) {

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BLOCK_DEVICE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasBlockDevice(machine),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}
