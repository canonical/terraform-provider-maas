package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASMachine_basic(t *testing.T) {
	vmHostID := os.Getenv("TF_ACC_VM_HOST_ID")
	testMachineName := acctest.RandomWithPrefix("tf-acc-ds-machine")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASMachineVMHostConfig(vmHostID, testMachineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "architecture"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "domain"),
					resource.TestCheckResourceAttr("data.maas_machine.test", "hostname", testMachineName),
					resource.TestCheckNoResourceAttr("data.maas_machine.test", "min_hw_kernel"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "pool"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "power_parameters"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "power_type"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "pxe_mac_address"),
					resource.TestCheckResourceAttr("data.maas_machine.test", "status", "Ready"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "zone"),
					resource.TestCheckResourceAttrSet("data.maas_machine.test", "block_devices.#"),
					testAccCheckNoBlockDeviceWithName("data.maas_machine.test", "test-volume-group-virtual-test"),
				),
			},
		},
	})
}

func testAccCheckNoBlockDeviceWithName(resourceName, deviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		countStr := rs.Primary.Attributes["block_devices.#"]

		count, _ := strconv.Atoi(countStr)
		for i := 0; i < count; i++ {
			name := rs.Primary.Attributes[fmt.Sprintf("block_devices.%d.name", i)]
			if name == deviceName {
				return fmt.Errorf("block_devices contains a virtual device %q but only physical devices should be present", deviceName)
			}
		}

		return nil
	}
}

func testAccDataSourceMAASMachineVMHostConfig(vmHostID, testMachineName string) string {
	return fmt.Sprintf(`
resource "maas_vm_host_machine" "test" {
  vm_host  = %q
  hostname = %q
}

# Create a virtual block device to verify that only physical block devices are returned
# by the data source
resource "maas_block_device" "test" {
  machine        = maas_vm_host_machine.test.id
  name           = "test-block-device"
  id_path        = "/dev/sda"
  size_gigabytes = 2
}

resource "maas_volume_group" "test" {
  machine       = maas_vm_host_machine.test.id
  name          = "test-volume-group"
  block_devices = [maas_block_device.test.id]
}

resource "maas_logical_volume" "test" {
  machine        = maas_vm_host_machine.test.id
  name           = "virtual-test"
  volume_group   = maas_volume_group.test.id
  size_gigabytes = 1
}

data "maas_machine" "test" {
  hostname   = maas_vm_host_machine.test.hostname
  depends_on = [maas_logical_volume.test]
}
`, vmHostID, testMachineName)
}
