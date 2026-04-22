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

func testAccMAASNetworkInterfacePhysical(fabricName string, name string, machine string, macAddress string, mtu int) string {
	return fmt.Sprintf(`
resource "maas_fabric" "default" {
    name = "%s"
}

data "maas_machine" "machine" {
    hostname = "%s"
}

data "maas_vlan" "default" {
    fabric = maas_fabric.default.id
    vlan   = 0
}

resource "maas_subnet" "test_subnet" {
    fabric       = maas_fabric.default.id
    vlan         = data.maas_vlan.default.id
    cidr         = "%s"
}

resource "maas_network_interface_physical" "test" {
    machine     = data.maas_machine.machine.id
    name        = "%s"
    mac_address = "%s"
    mtu         = %d
    tags        = ["tag1", "tag2"]
    vlan        = data.maas_vlan.default.id

    # When a physical interface is disconnected from a VLAN, MAAS automatically deletes the fabric
    # if the VLAN has no other interfaces or subnets attached. To prevent the fabric from being
    # deleted before the interface is disconnected, we add an explicit dependency on the subnet.
    # This ensures the subnet (and thus the fabric) remains until after the interface is
    # disconnected, allowing us to verify the disconnection without the fabric being removed
    # prematurely.
    # https://github.com/canonical/maas/commit/885021185340f740355faf13ad17b8fde5d8d285
    depends_on = [maas_subnet.test_subnet]
  }
`, fabricName, machine, testutils.GenerateRandomCIDR(), name, macAddress, mtu)
}

func TestAccResourceMAASNetworkInterfacePhysical_basic(t *testing.T) {
	var networkInterfacePhysical entity.NetworkInterface

	fabricName := acctest.RandomWithPrefix("tf-fab")

	name := fmt.Sprintf("tf-nic-eth-%d", acctest.RandIntRange(0, 999))

	machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	macAddress := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMAASNetworkInterfacePhysicalCheckExists("maas_network_interface_physical.test", &networkInterfacePhysical),
		resource.TestCheckResourceAttr("maas_network_interface_physical.test", "name", name),
		resource.TestCheckResourceAttr("maas_network_interface_physical.test", "mac_address", macAddress),
		resource.TestCheckResourceAttr("maas_network_interface_physical.test", "tags.#", "2"),
		resource.TestCheckResourceAttr("maas_network_interface_physical.test", "tags.0", "tag1"),
		resource.TestCheckResourceAttr("maas_network_interface_physical.test", "tags.1", "tag2"),
		resource.TestCheckResourceAttrPair("maas_network_interface_physical.test", "vlan", "data.maas_vlan.default", "id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_NETWORK_INTERFACE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASNetworkInterfacePhysicalDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				// Pass the dynamic fabric name into the config generator
				Config: testAccMAASNetworkInterfacePhysical(fabricName, name, machine, macAddress, 1500),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_physical.test", "mtu", "1500"))...),
			},
			// Test update
			{
				Config: testAccMAASNetworkInterfacePhysical(fabricName, name, machine, macAddress, 9000),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_physical.test", "mtu", "9000"))...),
			},
			// Test import
			{
				ResourceName:      "maas_network_interface_physical.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_network_interface_physical.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_network_interface_physical.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}

					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["machine"], rs.Primary.Attributes["id"]), nil
				},
			},
			// Test import by MAC Address
			{
				ResourceName:      "maas_network_interface_physical.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_network_interface_physical.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_network_interface_physical.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}

					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["machine"], rs.Primary.Attributes["mac_address"]), nil
				},
			},
		},
	})
}

func testAccMAASNetworkInterfacePhysicalCheckExists(rn string, networkInterfacePhysical *entity.NetworkInterface) resource.TestCheckFunc {
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

		gotNetworkInterfacePhysical, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err != nil {
			return fmt.Errorf("error getting network interface physical: %s", err)
		}

		*networkInterfacePhysical = *gotNetworkInterfacePhysical

		return nil
	}
}

func testAccCheckMAASNetworkInterfacePhysicalDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_network_interface_physical
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_network_interface_physical" {
			continue
		}

		// Retrieve our maas_network_interface_physical by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err == nil {
			// because this device is physical, we need to check it was *disconnected*, not destroyed
			if response != nil && response.VLAN.ID != 0 {
				return fmt.Errorf("MAAS Network interface physical (%s) still exists and is not disconnected from VLAN (%d)", rs.Primary.ID, response.VLAN.ID)
			}

			continue
		}

		// If the error is equivalent to 404 not found, the maas_network_interface_physical is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
