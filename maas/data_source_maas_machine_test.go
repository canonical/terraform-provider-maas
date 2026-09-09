package maas_test

import (
	"fmt"
	"os"
	"regexp"
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
					resource.TestCheckTypeSetElemAttr("data.maas_machine.test", "block_devices.1.tags.*", "test-tag"),
					testAccCheckNoBlockDeviceWithName("data.maas_machine.test", "test-volume-group-virtual-test"),
					// A lookup by system_id resolves to the same machine as the lookup by hostname
					resource.TestCheckResourceAttrPair("data.maas_machine.test", "system_id", "maas_vm_host_machine.test", "id"),
					resource.TestCheckResourceAttrPair("data.maas_machine.test_by_system_id", "id", "maas_vm_host_machine.test", "id"),
					resource.TestCheckResourceAttrPair("data.maas_machine.test_by_system_id", "system_id", "maas_vm_host_machine.test", "id"),
					resource.TestCheckResourceAttr("data.maas_machine.test_by_system_id", "hostname", testMachineName),
					resource.TestCheckResourceAttrPair("data.maas_machine.test_by_system_id", "pxe_mac_address", "data.maas_machine.test", "pxe_mac_address"),
					resource.TestCheckResourceAttrPair("data.maas_machine.test_by_system_id", "block_devices.#", "data.maas_machine.test", "block_devices.#"),
				),
			},
		},
	})
}

// The three identifiers are mutually exclusive, and at least one must be given.
func TestAccDataSourceMAASMachine_identifierExactlyOneOf(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: `
data "maas_machine" "test" {
  system_id = "abc123"
  hostname  = "tf-acc-ds-machine"
}
`,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
			{
				Config: `
data "maas_machine" "test" {
  system_id       = "abc123"
  pxe_mac_address = "52:54:00:89:f5:3e"
}
`,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
			{
				Config: `
data "maas_machine" "test" {}
`,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
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

resource "maas_block_device_tag" "test" {
  machine         = maas_vm_host_machine.test.id
  block_device_id = maas_block_device.test.id
  tags            = ["test-tag"]
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

# Use the same vm host machine for this test
data "maas_machine" "test_by_system_id" {
  system_id  = maas_vm_host_machine.test.id
  depends_on = [maas_logical_volume.test]
}
`, vmHostID, testMachineName)
}
