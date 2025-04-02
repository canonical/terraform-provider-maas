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

data "maas_fabric" "test" {
  name = "fabric-0"
}

data "maas_rack_controller" "test" {
  hostname = "maas-dev"
}

data "maas_vlan" "test" {
  vlan = "0"
  fabric = data.maas_fabric.test.id
}

resource "maas_subnet" "test" {
  cidr = "10.66.66.0/24"
  fabric = data.maas_fabric.test.id
  vlan = data.maas_vlan.test.id
  name = "subnet-66-66"
}

resource "maas_subnet_ip_range" "test" {
  subnet = maas_subnet.test.id
  start_ip = "10.66.66.1"
  end_ip = "10.66.66.254"
  type = "dynamic"
}

resource "maas_vlan_dhcp" "test" {
  fabric = data.maas_fabric.test.id
  vlan = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  ip_ranges = [maas_subnet_ip_range.test.id]
}

`
}