package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSubnet_basic(t *testing.T) {
	subnetName := acctest.RandomWithPrefix("test-subnet")
	cidr := testutils.GenerateRandomCIDR()
	gatewayIP := testutils.GetNetworkPrefixFromCIDR(cidr) + ".1"
	activeDiscovery := true
	allowDNS := true
	allowProxy := true
	managed := true

	changedSubnetName := acctest.RandomWithPrefix("test-subnet")
	changedCidr := testutils.GenerateRandomCIDR()
	changedGatewayIP := testutils.GetNetworkPrefixFromCIDR(changedCidr) + ".1"
	changedActiveDiscovery := false
	changedAllowDNS := false
	changedAllowProxy := false
	changedManaged := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASSubnetDestroy,
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccSubnetConfig(activeDiscovery, allowDNS, allowProxy, cidr, gatewayIP, managed, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASSubnetExists("maas_subnet.test"),
					resource.TestCheckResourceAttr("maas_subnet.test", "name", subnetName),
					resource.TestCheckResourceAttr("maas_subnet.test", "cidr", cidr),
					resource.TestCheckResourceAttr("maas_subnet.test", "gateway_ip", gatewayIP),
					resource.TestCheckResourceAttr("maas_subnet.test", "active_discovery", fmt.Sprintf("%t", activeDiscovery)),
					resource.TestCheckResourceAttr("maas_subnet.test", "allow_dns", fmt.Sprintf("%t", allowDNS)),
					resource.TestCheckResourceAttr("maas_subnet.test", "allow_proxy", fmt.Sprintf("%t", allowProxy)),
					resource.TestCheckResourceAttr("maas_subnet.test", "managed", fmt.Sprintf("%t", managed)),
				),
			},
			// Test update
			{
				Config: testAccSubnetConfig(changedActiveDiscovery, changedAllowDNS, changedAllowProxy, changedCidr, changedGatewayIP, changedManaged, changedSubnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASSubnetExists("maas_subnet.test"),
					resource.TestCheckResourceAttr("maas_subnet.test", "name", changedSubnetName),
					resource.TestCheckResourceAttr("maas_subnet.test", "cidr", changedCidr),
					resource.TestCheckResourceAttr("maas_subnet.test", "gateway_ip", changedGatewayIP),
					resource.TestCheckResourceAttr("maas_subnet.test", "active_discovery", fmt.Sprintf("%t", changedActiveDiscovery)),
					resource.TestCheckResourceAttr("maas_subnet.test", "allow_dns", fmt.Sprintf("%t", changedAllowDNS)),
					resource.TestCheckResourceAttr("maas_subnet.test", "allow_proxy", fmt.Sprintf("%t", changedAllowProxy)),
					resource.TestCheckResourceAttr("maas_subnet.test", "managed", fmt.Sprintf("%t", changedManaged)),
				),
			},
			// Test import
			{
				ResourceName:      "maas_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_subnet.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: maas_subnet.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}
					return rs.Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func testAccSubnetConfig(
	activeDiscovery bool,
	allowDNS bool,
	allowProxy bool,
	cidr string,
	gatewayIP string,
	managed bool,
	name string,
) string {
	return fmt.Sprintf(`
resource "maas_subnet" "test" {
	active_discovery = %t
	allow_dns		 = %t
	allow_proxy		 = %t
	cidr			 = %q
	dns_servers		 = ["8.8.8.8"]
	gateway_ip		 = %q
	managed			 = %t
	name			 = %q
	fabric			 = "0"
	vlan			 = "0"
}`, activeDiscovery, allowDNS, allowProxy, cidr, gatewayIP, managed, name)
}

func testAccCheckMAASSubnetExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.Subnet.Get(id)
		if err != nil {
			return fmt.Errorf("error getting subnet: %s", err)
		}

		return nil
	}
}

func testAccCheckMAASSubnetDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_subnet" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := conn.Subnet.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("Subnet group %s (%d) still exists.", response.Name, id)
			}
		}

		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
