package maas_test

import (
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestACCResourceMaasNetworkInterfaceBridge(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testutils.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestACCResourceMaasNetworkInterfaceBridgeConfig_basic,
			},
			{
				Config: TestACCResourceMaasNetworkInterfaceBridgeConfig_update,
				// ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "bridge_stp", "true"),
				),
			},
		},
	})
}

const TestACCResourceMaasNetworkInterfaceBridgeConfig_basic = `
resource "maas_network_interface_bridge" "test" {
	machine = "mq4s3r"
	name = "cloud-brmgmt"
	parent = "bond0.3342"
  }
  `

const TestACCResourceMaasNetworkInterfaceBridgeConfig_update = `
resource "maas_network_interface_bridge" "test" {
	machine = "mq4s3r"
	name = "cloud-brmgmt"
	parent = "bond0.3342"
	bridge_stp = true
  }
  `
