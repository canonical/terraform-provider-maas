package maas_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMAASVLANDHCP_basic(t *testing.T) {
	// Test variables
	fabricName := acctest.RandomWithPrefix("tf-basic")
	cidr := testutils.GenerateRandomCIDR()
	networkPrefix := testutils.GetNetworkPrefixFromCIDR(cidr)
	startIP, endIP := networkPrefix+".2", networkPrefix+".5"
	startIP2, endIP2 := networkPrefix+".6", networkPrefix+".10"
	rackController := os.Getenv("TF_ACC_RACK_CONTROLLER_HOSTNAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_RACK_CONTROLLER_HOSTNAME"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASVLANDHCPDestroy,
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccMAASVLANDHCPConfigBasic(fabricName, rackController, cidr, startIP, endIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASVLANDHCPExists("maas_vlan_dhcp.test", fabricName),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "vlan", "0"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test", "fabric", "maas_fabric.test", "id"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test", "primary_rack_controller", "data.maas_rack_controller.test", "id"),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "ip_ranges.#", "1"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test", "ip_ranges.0", "maas_subnet_ip_range.test", "id"),
				),
			},
			// Test destroy for just the VLAN DHCP resource.
			{
				Config: testAccMAASVLANDHCPConfigCore(fabricName, rackController, cidr, startIP, endIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASVLANDHCPAttrsUnsetWhenDHCPOff(),
				),
			},
			// Test update. Turn DHCP back on in the first step, then try to update in the second.
			{
				Config: testAccMAASVLANDHCPConfigBasic(fabricName, rackController, cidr, startIP, endIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASVLANDHCPExists("maas_vlan_dhcp.test", fabricName),
				),
			},
			{
				Config:      testAccMAASVLANDHCPConfigBasicUpdate(fabricName, rackController, cidr, startIP, endIP, startIP2, endIP2),
				ExpectError: regexp.MustCompile("Changing 'ip_ranges' from .* to .* is not allowed. Please recreate the resource."),
			},
		},
	})
}

func TestAccMAASVLANDHCP_wrongIPRange(t *testing.T) {
	// Test variables
	fabricName := acctest.RandomWithPrefix("tf-wrong-ip-range")
	cidr := testutils.GenerateRandomCIDR()
	networkPrefix := testutils.GetNetworkPrefixFromCIDR(cidr)
	startIP, endIP := networkPrefix+".2", networkPrefix+".5"
	rackController := os.Getenv("TF_ACC_RACK_CONTROLLER_HOSTNAME")
	fabricName2 := acctest.RandomWithPrefix("tf-wrong-ip-range-2")
	cidr2 := testutils.GenerateRandomCIDR()
	networkPrefix2 := testutils.GetNetworkPrefixFromCIDR(cidr2)
	startIP2, endIP2 := networkPrefix2+".2", networkPrefix2+".5"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_RACK_CONTROLLER_HOSTNAME"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASVLANDHCPDestroy,
		Steps: []resource.TestStep{
			// Test error on create.
			{
				Config:      testAccMAASVLANDHCPPConfigWrongIPRange(fabricName, fabricName2, rackController, cidr, cidr2, startIP, startIP2, endIP, endIP2),
				ExpectError: regexp.MustCompile("is not in the same VLAN as the VLAN DHCP resource."),
			},
		},
	})
}

func TestAccMAASVLANDHCP_subnet(t *testing.T) {
	// Test variables
	fabricName := acctest.RandomWithPrefix("tf-subnet")
	cidr := testutils.GenerateRandomCIDR()
	networkPrefix := testutils.GetNetworkPrefixFromCIDR(cidr)
	startIP, endIP := networkPrefix+".2", networkPrefix+".5"
	cidr2 := testutils.GenerateRandomCIDR()
	networkPrefix2 := testutils.GetNetworkPrefixFromCIDR(cidr2)
	startIP2, endIP2 := networkPrefix2+".2", networkPrefix2+".5"
	rackController := os.Getenv("TF_ACC_RACK_CONTROLLER_HOSTNAME")
	cidrForSubnetUpdate := testutils.GenerateRandomCIDR()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_RACK_CONTROLLER_HOSTNAME"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASVLANDHCPDestroy,
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccMAASVLANDHCPConfigSubnet(fabricName, rackController, cidr, startIP, endIP, cidr2, startIP2, endIP2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASVLANDHCPExists("maas_vlan_dhcp.test", fabricName),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "vlan", "0"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test", "fabric", "maas_fabric.test", "id"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test", "primary_rack_controller", "data.maas_rack_controller.test", "id"),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test", "subnets.#", "1"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test", "subnets.0", "maas_subnet.test_subnet", "id"),
				),
			},
			// Test update.
			{
				Config:      testAccMAASVLANDHCPConfigSubnetUpdate(fabricName, rackController, cidr, startIP, endIP, startIP2, endIP2, cidr2, cidrForSubnetUpdate),
				ExpectError: regexp.MustCompile("Changing 'subnets' from .* to .* is not allowed. Please recreate the resource."),
			},
		},
	})
}

func TestAccMAASVLANDHCP_relay(t *testing.T) {
	// Test variables
	fabricName := acctest.RandomWithPrefix("tf-relay")
	dummyFabricName := acctest.RandomWithPrefix("tf-dummy")
	cidr := testutils.GenerateRandomCIDR()
	networkPrefix := testutils.GetNetworkPrefixFromCIDR(cidr)
	startIP, endIP := networkPrefix+".2", networkPrefix+".5"
	cidr2 := testutils.GenerateRandomCIDR()
	networkPrefix2 := testutils.GetNetworkPrefixFromCIDR(cidr2)
	startIP2, endIP2 := networkPrefix2+".2", networkPrefix2+".5"
	rackController := os.Getenv("TF_ACC_RACK_CONTROLLER_HOSTNAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_RACK_CONTROLLER_HOSTNAME"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMAASVLANDHCPDestroy,
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccMAASVLANDHCPConfigRelay(fabricName, rackController, cidr, startIP, endIP, cidr2, startIP2, endIP2, dummyFabricName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMAASVLANDHCPExists("maas_vlan_dhcp.test_2", dummyFabricName),
					resource.TestCheckResourceAttr("maas_vlan_dhcp.test_2", "vlan", "0"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test_2", "fabric", "maas_fabric.dummy", "id"),
					resource.TestCheckResourceAttrPair("maas_vlan_dhcp.test_2", "relay_vlan", "data.maas_vlan.test", "id"),
				),
			},
		},
	})
}

func testAccCheckMAASVLANDHCPExists(n string, fabricName string) resource.TestCheckFunc {
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

		vlanVIDString := rs.Primary.Attributes["vlan"]

		vlanVID, err := strconv.Atoi(vlanVIDString)
		if err != nil {
			return fmt.Errorf("error converting vlan id to int: %s", err)
		}

		vlan, err := client.VLAN.Get(fabricID, vlanVID)
		if err != nil {
			return fmt.Errorf("error getting VLAN: %s", err)
		}

		if vlan.Fabric != fabricName {
			return fmt.Errorf("VLAN fabric name does not match, expected %s, got %s", fabricName, vlan.Fabric)
		}
		// All tests use the default untagged VLAN with VID = 0
		if vlan.VID != 0 {
			return fmt.Errorf("VLAN ID is not 0, got %d", vlan.ID)
		}

		if !vlan.DHCPOn {
			if vlan.RelayVLAN == nil {
				return fmt.Errorf("VLAN DHCP is not enabled and no relay VLAN set, VLAN DHCP does not exist")
			}
		}

		return nil
	}
}

func testAccCheckMAASVLANDHCPAttrsUnsetWhenDHCPOff() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the fabric Id from state
		rs, ok := s.RootModule().Resources["maas_fabric.test"]
		if !ok {
			return fmt.Errorf("fabric not found")
		}

		fabricID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error converting fabric id to int: %s", err)
		}
		// Get the vlan Id from state
		rs, ok = s.RootModule().Resources["data.maas_vlan.test"]
		if !ok {
			return fmt.Errorf("vlan not found")
		}

		vlanVID, err := strconv.Atoi(rs.Primary.Attributes["vlan"])
		if err != nil {
			return fmt.Errorf("error converting vlan id to int: %s", err)
		}

		// Check if the attributes are set as expected on the VLAN in MAAS
		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		vlan, err := client.VLAN.Get(fabricID, vlanVID)
		if err != nil {
			return fmt.Errorf("error getting VLAN from MAAS: %s", err)
		}

		if vlan.DHCPOn {
			return fmt.Errorf("VLAN DHCP is still enabled, expected it to be disabled")
		}

		if vlan.PrimaryRack != "" {
			return fmt.Errorf("VLAN primary rack controller is not nil, expected nil")
		}

		if vlan.SecondaryRack != "" {
			return fmt.Errorf("VLAN secondary rack controller is not nil, expected nil")
		}

		if vlan.RelayVLAN != nil {
			return fmt.Errorf("VLAN relay VLAN is not nil, expected nil")
		}

		return nil
	}
}

func testAccCheckMAASVLANDHCPDestroy(s *terraform.State) error {
	client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_vlan_dhcp" {
			continue
		}
		// Get the relevant IDs
		fabricID, vlanVID, err := maas.SplitStateIDIntoInts(rs.Primary.ID, "/")
		if err != nil {
			return err
		}

		// Check the VLAN no longer has DHCP enabled
		vlan, err := client.VLAN.Get(fabricID, vlanVID)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}

			return fmt.Errorf("error getting VLAN: %s", err)
		}

		if vlan.DHCPOn {
			return fmt.Errorf("VLAN with vid %d has DHCP still enabled", vlanVID)
		}
	}

	return nil
}

func testAccMAASVLANDHCPConfigCore(fabricID string, rackController string, cidr string, startIP string, endIP string) string {
	return fmt.Sprintf(`
resource "maas_fabric" "test" {
  name = %q
}

data "maas_rack_controller" "test" {
  hostname = %q
}

data "maas_vlan" "test" {
  vlan   = 0
  fabric = maas_fabric.test.id
}

resource "maas_subnet" "test" {
  cidr   = %q
  fabric = maas_fabric.test.id
  vlan   = data.maas_vlan.test.id
}

resource "maas_subnet_ip_range" "test" {
  subnet   = maas_subnet.test.id
  start_ip = %q
  end_ip   = %q
  type     = "dynamic"
}

`, fabricID, rackController, cidr, startIP, endIP)
}

func testAccMAASVLANDHCPConfigBasic(fabricID string, rackController string, cidr string, startIP string, endIP string) string {
	return fmt.Sprintf(`
%s
resource "maas_vlan_dhcp" "test" {
  fabric                  = maas_fabric.test.id
  vlan                    = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  ip_ranges               = [maas_subnet_ip_range.test.id]
}

`, testAccMAASVLANDHCPConfigCore(fabricID, rackController, cidr, startIP, endIP))
}

func testAccMAASVLANDHCPConfigBasicUpdate(fabricID string, rackController string, cidr string, startIP string, endIP string, startIP2 string, endIP2 string) string {
	return fmt.Sprintf(`
%s
resource "maas_subnet_ip_range" "test_2" {
  subnet   = maas_subnet.test.id
  start_ip = %q
  end_ip   = %q
  type     = "dynamic"
}
resource "maas_vlan_dhcp" "test" {
  fabric                  = maas_fabric.test.id
  vlan                    = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  ip_ranges               = [maas_subnet_ip_range.test.id, maas_subnet_ip_range.test_2.id]
}

`, testAccMAASVLANDHCPConfigCore(fabricID, rackController, cidr, startIP, endIP), startIP2, endIP2)
}

func testAccMAASVLANDHCPPConfigWrongIPRange(fabricID string, fabricID2 string, rackController string, cidr string, cidr2 string, startIP string, startIP2 string, endIP string, endIP2 string) string {
	return fmt.Sprintf(`
%s

# Create IP ranges in a separate fabric and VLAN to the VLAN where DHCP will be enabled.
##
resource "maas_fabric" "separate_fabric" {
  name = %q
}

data "maas_vlan" "separate_vlan" {
  vlan   = 0
  fabric = maas_fabric.separate_fabric.id
}

resource "maas_subnet" "separate_subnet" {
  cidr   = %q
  fabric = maas_fabric.separate_fabric.id
  vlan   = data.maas_vlan.separate_vlan.vlan
}

resource "maas_subnet_ip_range" "separate_ip_range" {
  subnet   = maas_subnet.separate_subnet.id
  start_ip = %q
  end_ip   = %q
  type     = "dynamic"
}

resource "maas_vlan_dhcp" "test" {
  fabric                  = maas_fabric.test.id
  vlan                    = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  ip_ranges               = [maas_subnet_ip_range.test.id, maas_subnet_ip_range.separate_ip_range.id]
}
`, testAccMAASVLANDHCPConfigCore(fabricID, rackController, cidr, startIP, endIP), fabricID2, cidr2, startIP2, endIP2)
}

func testAccMAASVLANDHCPConfigSubnet(fabricID string, rackController string, cidr string, startIP string, endIP string, cidr2 string, startIP2 string, endIP2 string) string {
	return fmt.Sprintf(`
%s
# Subnet needs ip ranges to be set to inform terraform about the dependency.
# Any other subnets defined earlier in this config aren't used by the VLAN DHCP resource.
## 
resource "maas_subnet" "test_subnet" {
  cidr   = %q
  fabric = maas_fabric.test.id
  vlan   = data.maas_vlan.test.vlan
  ip_ranges {
    start_ip = %q
    end_ip   = %q
    type     = "dynamic"
  }
}

resource "maas_vlan_dhcp" "test" {
  fabric                  = maas_fabric.test.id
  vlan                    = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  subnets                 = [maas_subnet.test_subnet.id]
}
`, testAccMAASVLANDHCPConfigCore(fabricID, rackController, cidr, startIP, endIP), cidr2, startIP2, endIP2)
}

func testAccMAASVLANDHCPConfigSubnetUpdate(fabricID string, rackController string, cidr string, startIP string, endIP string, startIP2 string, endIP2 string, cidr2 string, cidr3 string) string {
	return fmt.Sprintf(`
%s
resource "maas_subnet" "test_subnet" {
  cidr   = %q
  fabric = maas_fabric.test.id
  vlan   = data.maas_vlan.test.vlan
  ip_ranges {
    start_ip = %q
    end_ip   = %q
    type     = "dynamic"
  }
}

# New subnet to be added to the VLAN DHCP resource
resource "maas_subnet" "new_subnet" {
  cidr   = %q
  fabric = maas_fabric.test.id
  vlan   = data.maas_vlan.test.vlan
}

resource "maas_vlan_dhcp" "test" {
  fabric                  = maas_fabric.test.id
  vlan                    = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  subnets                 = [maas_subnet.test_subnet.id, maas_subnet.new_subnet.id]
}
`, testAccMAASVLANDHCPConfigCore(fabricID, rackController, cidr, startIP, endIP), cidr2, startIP2, endIP2, cidr3)
}

func testAccMAASVLANDHCPConfigRelay(fabricID string, rackController string, cidr string, startIP string, endIP string, cidr2 string, startIP2 string, endIP2 string, dummyFabricID string) string {
	return fmt.Sprintf(`
%s

resource "maas_fabric" "dummy" {
  name = %q
}

data "maas_vlan" "dummy" {
  vlan   = 0 # the default untagged vlan on all new fabrics.
  fabric = maas_fabric.dummy.id
}

resource "maas_subnet" "dummy" {
  cidr   = %q
  fabric = maas_fabric.dummy.id
  vlan   = data.maas_vlan.dummy.vlan
}

resource "maas_subnet_ip_range" "dummy" {
  subnet   = maas_subnet.dummy.id
  start_ip = %q
  end_ip   = %q
  type     = "dynamic"
}

resource "maas_vlan_dhcp" "test_2" {
  fabric     = maas_fabric.dummy.id
  vlan       = data.maas_vlan.dummy.vlan
  ip_ranges  = [maas_subnet_ip_range.dummy.id]
  relay_vlan = data.maas_vlan.test.id
}

`, testAccMAASVLANDHCPConfigBasic(fabricID, rackController, cidr, startIP, endIP), dummyFabricID, cidr2, startIP2, endIP2)
}
