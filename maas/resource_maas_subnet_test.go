package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASSubnet_basic(t *testing.T) {
	subnetName := acctest.RandomWithPrefix("test-subnet")
	subnetNameMod := acctest.RandomWithPrefix("test-subnet-mod")
	subnetAttrName := "maas_subnet.test_subnet"
	fabricName := acctest.RandomWithPrefix("test-fabric")
	cidr := testutils.GenerateRandomCIDR()
	gateway := testutils.GetNetworkPrefixFromCIDR(cidr) + ".1"
	gatewayMod := testutils.GetNetworkPrefixFromCIDR(cidr) + ".254"
	vlan := "0" // Use default VLAN (VID 0) of the fabric
	dnsServer1 := "8.8.8.8"
	dnsServer2 := "8.8.4.4"
	dnsServer1Mod := "1.1.1.1"
	dnsServer2Mod := "1.0.0.1"
	ipRangeStart := testutils.GetNetworkPrefixFromCIDR(cidr) + ".10"
	ipRangeEnd := testutils.GetNetworkPrefixFromCIDR(cidr) + ".100"
	ipRangeStartMod := testutils.GetNetworkPrefixFromCIDR(cidr) + ".10"
	ipRangeEndMod := testutils.GetNetworkPrefixFromCIDR(cidr) + ".50"
	ipRangeType := "dynamic"
	ipRangeComment := "test-dynamic-range"
	ipRangeCommentMod := "modified-dynamic-range"
	allowDNS := true
	allowDNSMod := false
	allowProxy := true
	allowProxyMod := false
	rdnsMode := 2
	rdnsModeMod := 1

	// Check functions for initial creation
	checks := []resource.TestCheckFunc{
		testAccMAASSubnetCheckExists(subnetAttrName),
		resource.TestCheckResourceAttr(subnetAttrName, "cidr", cidr),
		resource.TestCheckResourceAttr(subnetAttrName, "name", subnetName),
		resource.TestCheckResourceAttrPair(subnetAttrName, "fabric", "maas_fabric.test_fabric", "id"),
		resource.TestCheckResourceAttr(subnetAttrName, "vlan", vlan),
		resource.TestCheckResourceAttr(subnetAttrName, "gateway_ip", gateway),
		resource.TestCheckResourceAttr(subnetAttrName, "dns_servers.0", dnsServer1),
		resource.TestCheckResourceAttr(subnetAttrName, "dns_servers.1", dnsServer2),
		resource.TestCheckResourceAttr(subnetAttrName, "allow_dns", fmt.Sprintf("%t", allowDNS)),
		resource.TestCheckResourceAttr(subnetAttrName, "allow_proxy", fmt.Sprintf("%t", allowProxy)),
		resource.TestCheckResourceAttr(subnetAttrName, "rdns_mode", fmt.Sprintf("%d", rdnsMode)),
		resource.TestCheckResourceAttr(subnetAttrName, "ip_ranges.#", "1"),
	}

	// Check functions for modified values
	checksMod := []resource.TestCheckFunc{
		testAccMAASSubnetCheckExists(subnetAttrName),
		resource.TestCheckResourceAttr(subnetAttrName, "cidr", cidr),
		resource.TestCheckResourceAttr(subnetAttrName, "name", subnetNameMod),
		resource.TestCheckResourceAttrPair(subnetAttrName, "fabric", "maas_fabric.test_fabric", "id"),
		resource.TestCheckResourceAttr(subnetAttrName, "vlan", vlan),
		resource.TestCheckResourceAttr(subnetAttrName, "gateway_ip", gatewayMod),
		resource.TestCheckResourceAttr(subnetAttrName, "dns_servers.0", dnsServer1Mod),
		resource.TestCheckResourceAttr(subnetAttrName, "dns_servers.1", dnsServer2Mod),
		resource.TestCheckResourceAttr(subnetAttrName, "allow_dns", fmt.Sprintf("%t", allowDNSMod)),
		resource.TestCheckResourceAttr(subnetAttrName, "allow_proxy", fmt.Sprintf("%t", allowProxyMod)),
		resource.TestCheckResourceAttr(subnetAttrName, "rdns_mode", fmt.Sprintf("%d", rdnsModeMod)),
		resource.TestCheckResourceAttr(subnetAttrName, "ip_ranges.#", "1"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASSubnetDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccSubnetExampleResource(
					cidr,
					subnetName,
					fabricName,
					vlan,
					gateway,
					dnsServer1,
					dnsServer2,
					allowDNS,
					allowProxy,
					rdnsMode,
					ipRangeType,
					ipRangeStart,
					ipRangeEnd,
					ipRangeComment,
				),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
			// Test update
			{
				Config: testAccSubnetExampleResource(
					cidr,
					subnetNameMod,
					fabricName,
					vlan,
					gatewayMod,
					dnsServer1Mod,
					dnsServer2Mod,
					allowDNSMod,
					allowProxyMod,
					rdnsModeMod,
					ipRangeType,
					ipRangeStartMod,
					ipRangeEndMod,
					ipRangeCommentMod,
				),
				Check: resource.ComposeTestCheckFunc(checksMod...),
			},
			// Test removal of ip_ranges
			{
				Config: testAccSubnetExampleResourceNoIPRanges(
					cidr,
					subnetNameMod,
					fabricName,
					vlan,
					gatewayMod,
					dnsServer1Mod,
					dnsServer2Mod,
					allowDNSMod,
					allowProxyMod,
					rdnsModeMod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASSubnetCheckExists(subnetAttrName),
					resource.TestCheckResourceAttr(subnetAttrName, "ip_ranges.#", "0"),
					testAccMAASSubnetCheckNoIPRanges(subnetAttrName),
				),
			},
			// Test import by ID
			{
				ResourceName:            subnetAttrName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ip_ranges"}, // IP ranges are not imported, this should be done using `resourceMAASSubnetIPRange`

			},
			// Test import by CIDR
			{
				ResourceName:            subnetAttrName,
				ImportState:             true,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{"ip_ranges"}, // IP ranges are not imported, this should be done using `resourceMAASSubnetIPRange`
				ImportStateId:           cidr,
			},
		},
	})
}

// Check if the subnet specified actually exists in MAAS
func testAccMAASSubnetCheckExists(rn string) resource.TestCheckFunc {
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

// Check that no IP ranges exist for the subnet in MAAS
func testAccMAASSubnetCheckNoIPRanges(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		subnetID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Get all IP ranges and check if any belong to this subnet
		ipRanges, err := conn.IPRanges.Get()
		if err != nil {
			return fmt.Errorf("error getting IP ranges: %s", err)
		}

		for _, ipr := range ipRanges {
			if ipr.Subnet.ID == subnetID {
				return fmt.Errorf("expected no IP ranges for subnet %d, but found IP range %d (type: %s, %s-%s)",
					subnetID, ipr.ID, ipr.Type, ipr.StartIP, ipr.EndIP)
			}
		}

		return nil
	}
}

// An example resource configuration for a subnet with all parameters
func testAccSubnetExampleResource(
	cidr string,
	name string,
	fabricName string,
	vlan string,
	gateway string,
	dnsServer1 string,
	dnsServer2 string,
	allowDNS bool,
	allowProxy bool,
	rdnsMode int,
	ipRangeType string,
	ipRangeStart string,
	ipRangeEnd string,
	ipRangeComment string,
) string {
	return fmt.Sprintf(`
resource "maas_fabric" "test_fabric" {
  name = %q
}

resource "maas_subnet" "test_subnet" {
  cidr        = %q
  name        = %q
  fabric      = maas_fabric.test_fabric.id
  vlan        = %q
  gateway_ip  = %q
  dns_servers = [%q, %q]
  allow_dns   = %t
  allow_proxy = %t
  rdns_mode   = %d

  ip_ranges {
    type     = %q
    start_ip = %q
    end_ip   = %q
    comment  = %q
  }
}
	`, fabricName, cidr, name, vlan, gateway, dnsServer1, dnsServer2, allowDNS, allowProxy, rdnsMode,
		ipRangeType, ipRangeStart, ipRangeEnd, ipRangeComment)
}

// An example resource configuration for a subnet without IP ranges
func testAccSubnetExampleResourceNoIPRanges(
	cidr string,
	name string,
	fabricName string,
	vlan string,
	gateway string,
	dnsServer1 string,
	dnsServer2 string,
	allowDNS bool,
	allowProxy bool,
	rdnsMode int,
) string {
	return fmt.Sprintf(`
resource "maas_fabric" "test_fabric" {
  name = %q
}

resource "maas_subnet" "test_subnet" {
  cidr        = %q
  name        = %q
  fabric      = maas_fabric.test_fabric.id
  vlan        = %q
  gateway_ip  = %q
  dns_servers = [%q, %q]
  allow_dns   = %t
  allow_proxy = %t
  rdns_mode   = %d
}
	`, fabricName, cidr, name, vlan, gateway, dnsServer1, dnsServer2, allowDNS, allowProxy, rdnsMode)
}

func testAccCheckMAASSubnetDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_subnet
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_subnet" {
			continue
		}

		// retrieve the maas_subnet by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Check if the subnet exists
		var exists bool

		response, err := conn.Subnet.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				exists = true
			} else {
				return fmt.Errorf("unexpected response when checking for subnet existence: %#v", response)
			}
		}

		if exists {
			return fmt.Errorf("MAAS %s (%s) still exists.", rs.Type, rs.Primary.ID)
		}

		// If the error is equivalent to 404 not found, the maas_subnet is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
