package maas_test

import (
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const TestAccResourceMaasNetworkInterfaceBridgeConfig_basic = `
resource "maas_network_interface_bridge" "test" {
	machine = "mq4s3r"
	name = "cloud-brmgmt"
	parent = "bond0.3342"
  }
`

const TestAccResourceMaasNetworkInterfaceBridgeConfig_update = `
resource "maas_network_interface_bridge" "test" {
	machine = "mq4s3r"
	name = "cloud-brmgmt"
	parent = "bond0.3342"
	bridge_stp = true
  }
`

func TestAccResourceMaasNetworkInterfaceBridge_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testutils.TestAccProviders,
		CheckDestroy: nil,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: TestAccResourceMaasNetworkInterfaceBridgeConfig_basic,
			},
			{
				Config: TestAccResourceMaasNetworkInterfaceBridgeConfig_update,
				// ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "bridge_stp", "true"),
				),
			},
		},
	})
}
