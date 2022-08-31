package maas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceMaasVmHost(t *testing.T) {

	tf := fmt.Sprintf(testAccResourceMaasVmHostConfig)
	resourceName := "maas_vm_host.kvm"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          nil,
		ProviderFactories: providerFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "lxd"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
				),
			},
		},
	})
}

const testAccResourceMaasVmHostConfig = `
resource "maas_vm_host" "kvm" {
  type = "lxd"
  power_address = "10.10.10.8"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
}
`
