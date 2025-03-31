package maas_test

import (
	"testing"

	"terraform-provider-maas/maas/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccMaasVlanDHCP_basic(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccVlanDHCPConfigBasic(),
				// Check: resource.ComposeTestCheckFunc(
				// 	testAccCheckMaasVlanDHCPExists("maas_vlan_dhcp.test", vlanID, subnetID),
				// 	resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "vlan", strconv.Itoa(vlanID)),
				// 	resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "subnets.#", "1"),
				// 	resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "subnets.0", strconv.Itoa(subnetID)),
				// 	resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "primary_rack_controller", ""),
				// ),
			},
		},
	})
}
		
func testAccVlanDHCPConfigBasic() string {
return`
data "maas_fabric" "fabric" {
	name = "fabric-0"
}

data "maas_vlan" "vlan" {
    fabric = "fabric-0"
    vlan    = 0
}

data "maas_rack_controller" "controller_primary" {
    hostname = "maas-dev"
}

resource "maas_subnet" "subnet_1" {
  fabric = data.maas_fabric.fabric.id
  vlan   = data.maas_vlan.vlan.id
  name   = "test_subnet"

  cidr        = "10.66.66.0/24"
  gateway_ip  = "10.66.66.1"
  dns_servers = [
    "1.1.1.1",
  ]
}\

resource "maas_subnet_ip_range" "dynamic_ip_range_1_1" {
  subnet   = maas_subnet.subnet_1.id
  type     = "dynamic"
  start_ip = "10.66.66.2"
  end_ip   = "10.66.66.60"
}

resource "maas_subnet_ip_range" "dynamic_ip_range_1_2" {
  subnet   = maas_subnet.subnet_1.id
  type     = "dynamic"
  start_ip = "10.66.66.61"
  end_ip   = "10.66.66.120"
}

resource "maas_vlan_dhcp" "dhcp" {
    vlan = data.maas_vlan.vlan.id
	fabric = data.maas_fabric.fabric.id
    
    ip_ranges = [
        maas_subnet_ip_range.dynamic_ip_range_1_1.id,
        maas_subnet_ip_range.dynamic_ip_range_1_2.id,
    ]


    primary_rack_controller = data.maas_rack_controller.controller_primary.id
}
`
}