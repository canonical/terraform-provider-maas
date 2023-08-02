package maas_test

import (
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const TestACCResourceMaasNetworkInterfaceVlanConfig_basic = `
resource "maas_network_interface_vlan" "test" {
	machine = "mq4s3r"
	parent = "bond0"
	vlan = "3342"
	fabric = "fabric-hydc"
	accept_ra = false
	mtu = 9000
  }
  `

const TestACCResourceMaasNetworkInterfaceVlanConfig_update = `
resource "maas_network_interface_vlan" "test" {
	machine = "mq4s3r"
	parent = "bond0"
	vlan = "3342"
	fabric = "fabric-hydc"
	accept_ra = true
	mtu = 9001
  }
  `

func TestACCResourceMaasNetworkInterfaceVlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testutils.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestACCResourceMaasNetworkInterfaceVlanConfig_basic,
			},
			{
				Config: TestACCResourceMaasNetworkInterfaceVlanConfig_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "mtu", "9001"),
				),
			},
		},
	})
}
