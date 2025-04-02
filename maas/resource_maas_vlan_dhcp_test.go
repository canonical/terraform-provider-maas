package maas_test

import (
	"testing"
	"strconv"
	"fmt"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccMaasVlanDHCP_basic(t *testing.T) {
	vlanID := 0
	fabricID := "0"
	cidr := testutils.GenerateRandomCidr()
	networkPrefix := testutils.GetNetworkPrefixFromCidr(cidr)
	startIP, endIP := networkPrefix + ".2", networkPrefix + ".5"
	rackController := "maas-dev"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASVLANDHCPCheckDestroy,
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccVlanDHCPConfigBasic(fabricID, rackController, vlanID, cidr, startIP, endIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaasVlanDHCPExists("maas_vlan_dhcp.test", vlanID),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "vlan", strconv.Itoa(vlanID)),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "fabric", fabricID),
					resource.TestCheckResourceAttrSet("maas_vlan_dhcp.test", "primary_rack_controller"),
				),
			},
		},
	})
}

func testAccCheckMaasVlanDHCPExists(n string, vlanID int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", n, s.RootModule().Resources)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}
		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		fabricIDString := rs.Primary.Attributes["fabric"]
		fabricID, err := strconv.Atoi(fabricIDString)
		if err != nil {
			return fmt.Errorf("error converting fabric id to int: %s", err)
		}
		vlan, err := client.VLAN.Get(fabricID, vlanID)
		if err != nil {
			return fmt.Errorf("error getting VLAN: %s", err)
		}
		if vlan.VID != vlanID {
			return fmt.Errorf("VLAN id mismatch: %d != %d", vlan.ID, vlanID)
		}
		if !vlan.DHCPOn {
			return fmt.Errorf("VLAN DHCP is not enabled, resource failed to turn on DHCP")
		}
		return nil
	}
}

func testAccCheckMAASVLANDHCPCheckDestroy(s *terraform.State) error {
	client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_vlan_dhcp" {
			continue
		}
		// Get the relevant IDs
		vlanID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		fabricIDString := rs.Primary.Attributes["fabric"]
		fabricID, err := strconv.Atoi(fabricIDString)
		if err != nil {
			return err
		}
		// Check the VLAN no longer has DHCP enabled
		vlan, err := client.VLAN.Get(fabricID, vlanID)
		if err != nil {
			return err
		}
		if vlan.DHCPOn {
			return fmt.Errorf("VLAN with vid %d has DHCP still enabled", vlanID)
		}
	}
	return nil
}
	
func testAccVlanDHCPConfigBasic(fabricID string, rackController string, vlanID int, cidr string, startIP string, endIP string) string {
	return fmt.Sprintf(`
data "maas_fabric" "test" {
  name = %q
}

data "maas_rack_controller" "test" {
  hostname = %q
}

data "maas_vlan" "test" {
  vlan = %d
  fabric = data.maas_fabric.test.id
}

resource "maas_subnet" "test" {
  cidr = %q
  fabric = data.maas_fabric.test.id
  vlan = data.maas_vlan.test.id
  name = "subnet-66-66"
}

resource "maas_subnet_ip_range" "test" {
  subnet = maas_subnet.test.id
  start_ip = %q
  end_ip = %q
  type = "dynamic"
}

resource "maas_vlan_dhcp" "test" {
  fabric = data.maas_fabric.test.id
  vlan = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  ip_ranges = [maas_subnet_ip_range.test.id]
}

`, fabricID, rackController, vlanID, cidr, startIP, endIP)
}
