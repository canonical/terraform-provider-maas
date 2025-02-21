package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccMaasNetworkInterfaceBridge(name string, machine string, mac_address string, mac_address_phys string, mtu int) string {
	return fmt.Sprintf(`
resource "maas_fabric" "default" {
	name = "tf-fabric-bridge"
}

data "maas_machine" "machine" {
	hostname = "%s"
}

data "maas_vlan" "default" {
	fabric = maas_fabric.default.id
	vlan   = 0
}

resource "maas_network_interface_physical" "nic1" {
	machine     = data.maas_machine.machine.id
	mac_address = "%s"
	name        = "ethbr"
	vlan        = data.maas_vlan.default.id
}

resource "maas_network_interface_bridge" "test" {
	machine    = data.maas_machine.machine.id
	name       = "%s"
	accept_ra   = false
	bridge_fd   = 42
	bridge_stp  = true
	bridge_type = "standard"
	mac_address = "%s"
	mtu         = %d
	parent      = maas_network_interface_physical.nic1.name
	tags        = ["tag1", "tag2"]
	vlan        = data.maas_vlan.default.id
  }
`, machine, mac_address_phys, name, mac_address, mtu)
}

func TestAccResourceMaasNetworkInterfaceBridge_basic(t *testing.T) {

	var networkInterfaceBridge entity.NetworkInterface
	name := fmt.Sprintf("tf-nic-br-%d", acctest.RandIntRange(0, 9))
	machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	mac_address := testutils.RandomMAC()
	mac_address_phys := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMaasNetworkInterfaceBridgeCheckExists("maas_network_interface_bridge.test", &networkInterfaceBridge),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "name", name),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "accept_ra", "false"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "bridge_fd", "42"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "bridge_stp", "true"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "bridge_type", "standard"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "mac_address", mac_address),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "parent", "ethbr"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "tags.#", "2"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "tags.0", "tag1"),
		resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "tags.1", "tag2"),
		resource.TestCheckResourceAttrPair("maas_network_interface_bridge.test", "vlan", "data.maas_vlan.default", "id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_NETWORK_INTERFACE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasNetworkInterfaceBridgeDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasNetworkInterfaceBridge(name, machine, mac_address, mac_address_phys, 1500),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "mtu", "1500"))...),
			},
			// Test update
			{
				Config: testAccMaasNetworkInterfaceBridge(name, machine, mac_address, mac_address_phys, 9000),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_bridge.test", "mtu", "9000"))...),
			},
			// Test import
			{
				ResourceName:      "maas_network_interface_bridge.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_network_interface_bridge.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_network_interface_bridge.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["machine"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func testAccMaasNetworkInterfaceBridgeCheckExists(rn string, networkInterfaceBridge *entity.NetworkInterface) resource.TestCheckFunc {
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
		gotNetworkInterfaceBridge, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err != nil {
			return fmt.Errorf("error getting network interface bridge: %s", err)
		}

		*networkInterfaceBridge = *gotNetworkInterfaceBridge

		return nil
	}
}

func testAccCheckMaasNetworkInterfaceBridgeDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_network_interface_bridge
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_network_interface_bridge" {
			continue
		}

		// Retrieve our maas_network_interface_bridge by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		response, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Network interface bridge (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_network_interface_bridge is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
