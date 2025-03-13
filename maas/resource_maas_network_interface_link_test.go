package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccMaasNetworkInterfaceLink(machine string, cidr string, gateway string, ip string, mac_address string) string {
	return fmt.Sprintf(`
data "maas_machine" "machine" {
  hostname = "%s"
}

resource "maas_fabric" "test" {
  name = "tf-fabric-link"
}

data "maas_vlan" "default" {
  fabric = maas_fabric.test.id
  vlan   = 0
}

resource "maas_vlan" "test" {
  fabric = maas_fabric.test.id
  vid    = 20
}

resource "maas_network_interface_physical" "test" {
  machine     = data.maas_machine.machine.id
  mac_address = "%s"
  name        = "tf0"
  vlan        = data.maas_vlan.default.id
}

resource "maas_network_interface_vlan" "test" {
  machine = data.maas_machine.machine.id
  fabric  = maas_fabric.test.id
  parent  = maas_network_interface_physical.test.name
  vlan    = maas_vlan.test.id
}

resource "maas_network_interface_bridge" "test" {
  machine = data.maas_machine.machine.id
  parent  = maas_network_interface_vlan.test.name
  vlan    = maas_vlan.test.id
  name    = "tfbr"
}

resource "maas_subnet" "test" {
  cidr       = "%s"
  fabric     = maas_fabric.test.id
  vlan       = maas_vlan.test.vid
  name       = "tf_subnet"
  gateway_ip = "%s"
}

resource "maas_network_interface_link" "test" {
  machine           = data.maas_machine.machine.id
  network_interface = maas_network_interface_bridge.test.id
  subnet            = maas_subnet.test.cidr
  mode              = "STATIC"
  ip_address        = "%s"
  default_gateway   = true
}
`, machine, mac_address, cidr, gateway, ip)
}

func testAccMaasNetworkInterfaceLinkDevice(macAddress string, randomName string, cidr string, gateway string, ip string) string {
	return fmt.Sprintf(`
resource "maas_device" "test" {
  hostname    = %q
  network_interfaces {
    mac_address = %q
  }
  depends_on = [maas_fabric.test]
}

resource "maas_fabric" "test" {
  name = %q
}

resource "maas_subnet" "test" {
  cidr       = %q
  name       = %q
  fabric     = maas_fabric.test.id
  gateway_ip = %q
}

resource "maas_network_interface_link" "first" {
  device            = maas_device.test.id
  network_interface = tolist(maas_device.test.network_interfaces)[0].id
  subnet            = maas_subnet.test.cidr
  mode              = "STATIC"
  ip_address        = %q
}
`, randomName, macAddress, randomName, cidr, randomName, gateway, ip)
}

func TestAccResourceMaasNetworkInterfaceLink_device(t *testing.T) {
	macAddress := testutils.RandomMAC()
	randomName := acctest.RandomWithPrefix("tf-test")
	cidr := "10.77.77.0/24"
	gateway := "10.77.77.1"
	ipAddress := "10.77.77.42"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasNetworkInterfaceLinkDevice(macAddress, randomName, cidr, gateway, ipAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccMaasNetworkInterfaceLinkCheckExists("maas_network_interface_link.first", "device"),
					resource.TestCheckResourceAttr("maas_network_interface_link.first", "ip_address", ipAddress),
					resource.TestCheckResourceAttr("maas_network_interface_link.first", "mode", "STATIC"),
					resource.TestCheckResourceAttr("maas_network_interface_link.first", "subnet", cidr),
					resource.TestCheckResourceAttrPair("maas_network_interface_link.first", "device", "maas_device.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceMaasNetworkInterfaceLink_basic(t *testing.T) {

	machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	cidr := "30.30.30.0/24"
	gateway := "30.30.30.1"
	mac_address := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMaasNetworkInterfaceLinkCheckExists("maas_network_interface_link.test", "machine"),
		resource.TestCheckResourceAttr("maas_network_interface_link.test", "subnet", cidr),
		resource.TestCheckResourceAttr("maas_network_interface_link.test", "mode", "STATIC"),
		resource.TestCheckResourceAttr("maas_network_interface_link.test", "default_gateway", "true"),
		resource.TestCheckResourceAttrPair("maas_network_interface_link.test", "machine", "data.maas_machine.machine", "id"),
		resource.TestCheckResourceAttrPair("maas_network_interface_link.test", "network_interface", "maas_network_interface_bridge.test", "id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_NETWORK_INTERFACE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasNetworkInterfaceLink(machine, cidr, gateway, "30.30.30.2", mac_address),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_link.test", "ip_address", "30.30.30.2"))...),
			},
			// Test update
			{
				Config: testAccMaasNetworkInterfaceLink(machine, cidr, gateway, "30.30.30.3", mac_address),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_link.test", "ip_address", "30.30.30.3"))...),
			},
		},
	})
}

func testAccMaasNetworkInterfaceLinkCheckExists(rn string, nodeType string) resource.TestCheckFunc {
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
		networkInterfaceID, err := strconv.Atoi(rs.Primary.Attributes["network_interface"])
		if err != nil {
			return err
		}

		gotNetworkInterface, err := conn.NetworkInterface.Get(rs.Primary.Attributes[nodeType], networkInterfaceID)
		if err != nil {
			return fmt.Errorf("error getting network interface: %s", err)
		}

		for _, link := range gotNetworkInterface.Links {
			if id == link.ID {
				return nil
			}
		}

		return fmt.Errorf("link with id: %v not found in the network interface links", id)
	}
}
