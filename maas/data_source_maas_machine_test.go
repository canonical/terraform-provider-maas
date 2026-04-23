package maas_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				),
			},
		},
	})
}

func testAccDataSourceMAASMachineVMHostConfig(vmHostID, testMachineName string) string {
	return fmt.Sprintf(`
resource "maas_vm_host_machine" "test" {
  vm_host  = %q
  hostname = %q
}

data "maas_machine" "test" {
  hostname = maas_vm_host_machine.test.hostname
}
`, vmHostID, testMachineName)
}
