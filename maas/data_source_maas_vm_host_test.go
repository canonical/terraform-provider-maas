package maas_test

import (
	"fmt"
	"os"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var baseCheckFunctions = []resource.TestCheckFunc{
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "type", "maas_vm_host.test", "type"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "name", "maas_vm_host.test", "name"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "cpu_over_commit_ratio", "maas_vm_host.test", "cpu_over_commit_ratio"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "default_macvlan_mode", "maas_vm_host.test", "default_macvlan_mode"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "memory_over_commit_ratio", "maas_vm_host.test", "memory_over_commit_ratio"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "pool", "maas_vm_host.test", "pool"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "power_address", "maas_vm_host.test", "power_address"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "resources_cores_total", "maas_vm_host.test", "resources_cores_total"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "resources_local_storage_total", "maas_vm_host.test", "resources_local_storage_total"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "resources_memory_total", "maas_vm_host.test", "resources_memory_total"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "tags.#", "maas_vm_host.test", "tags.#"),
	resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "zone", "maas_vm_host.test", "zone"),
}

func testAccDataSourceMAASVMHost(t *testing.T, vmHostType string, additionalChecks ...resource.TestCheckFunc) {
	vmHostID := os.Getenv("TF_ACC_VM_HOST_ID")
	vmHostMachineName := acctest.RandomWithPrefix(fmt.Sprintf("tf-data-source-vm-host-%s", vmHostType))

	checks := make([]resource.TestCheckFunc, 0, len(baseCheckFunctions)+len(additionalChecks)+1)
	checks = append(checks, baseCheckFunctions...)
	checks = append(checks, checkMAASVMHostExists(t, "maas_vm_host.test"))
	checks = append(checks, additionalChecks...)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASVMHostDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASVMHostConfig(vmHostID, vmHostMachineName, vmHostType),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDataSourceMAASVMHost_lxd(t *testing.T) {
	vmHostType := "lxd"

	testAccDataSourceMAASVMHost(t, vmHostType,
		checkMAASVMHostExists(t, "maas_vm_host.test"),
		resource.TestCheckResourceAttr("data.maas_vm_host.test", "type", vmHostType),
		resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "project", "maas_vm_host.test", "project"),
		resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "certificate", "maas_vm_host.test", "certificate"),
		resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "key", "maas_vm_host.test", "key"),
		resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "project", "maas_vm_host.test", "project"),
		resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "password", "maas_vm_host.test", "password"))
}

func TestAccDataSourceMAASVMHost_virsh(t *testing.T) {
	vmHostType := "virsh"

	testAccDataSourceMAASVMHost(t, vmHostType,
		resource.TestCheckResourceAttr("data.maas_vm_host.test", "type", vmHostType),
		resource.TestCheckResourceAttrPair("data.maas_vm_host.test", "power_pass", "maas_vm_host.test", "power_pass"),
	)
}

func testAccDataSourceMAASVMHostConfig(vmHost string, vmHostMachineName string, hostType string) string {
	return fmt.Sprintf(`
%s

data "maas_vm_host" "test" {
  name = maas_vm_host.test.name
}
`, testAccMAASVMHostDeployParamsConfig(vmHost, vmHostMachineName, hostType))
}
