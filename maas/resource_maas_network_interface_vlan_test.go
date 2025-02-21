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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccMaasNetworkInterfaceVLAN(machine string, mac_adrress_phys string, mtu int) string {
	return fmt.Sprintf(`
resource "maas_fabric" "default" {
	name = "tf-fabric-vlan"
}

data "maas_machine" "machine" {
	hostname = "%s"
}

data "maas_vlan" "default" {
	fabric = maas_fabric.default.id
	vlan   = 0
}

resource "maas_vlan" "tf_vlan" {
	fabric = maas_fabric.default.id
	vid    = 42
	name   = "tf-vlan-42"
}

resource "maas_network_interface_physical" "nic1" {
	machine     = data.maas_machine.machine.id
	mac_address = "%s"
	name        = "bond0"
	vlan        = data.maas_vlan.default.id
}

resource "maas_network_interface_vlan" "test" {
	machine   = data.maas_machine.machine.id
	accept_ra = false
	fabric    = maas_fabric.default.id
	mtu       = %d
	parent    = maas_network_interface_physical.nic1.name
	tags      = ["tag1", "tag2"]
	vlan      = maas_vlan.tf_vlan.id
}
  `, machine, mac_adrress_phys, mtu)
}

func TestAccResourceMaasNetworkInterfaceVLAN_basic(t *testing.T) {

	var networkInterfaceVLAN entity.NetworkInterface
	machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	mac_adrress_phys := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMaasNetworkInterfaceVLANCheckExists("maas_network_interface_vlan.test", &networkInterfaceVLAN),
		resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "accept_ra", "false"),
		resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "parent", "bond0"),
		resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "tags.#", "2"),
		resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "tags.0", "tag1"),
		resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "tags.1", "tag2"),
		resource.TestCheckResourceAttrPair("maas_network_interface_vlan.test", "vlan", "maas_vlan.tf_vlan", "id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_NETWORK_INTERFACE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasNetworkInterfaceVLANDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasNetworkInterfaceVLAN(machine, mac_adrress_phys, 1500),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "mtu", "1500"))...),
			},
			// Test update
			{
				Config: testAccMaasNetworkInterfaceVLAN(machine, mac_adrress_phys, 9000),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_vlan.test", "mtu", "9000"))...),
			},
			// Test import
			{
				ResourceName:      "maas_network_interface_vlan.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_network_interface_vlan.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_network_interface_vlan.test")
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

func testAccMaasNetworkInterfaceVLANCheckExists(rn string, networkInterfaceVLAN *entity.NetworkInterface) resource.TestCheckFunc {
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
		gotNetworkInterfaceVLAN, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err != nil {
			return fmt.Errorf("error getting network interface VLAN: %s", err)
		}

		*networkInterfaceVLAN = *gotNetworkInterfaceVLAN

		return nil
	}
}

func testAccCheckMaasNetworkInterfaceVLANDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_network_interface_vlan
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_network_interface_vlan" {
			continue
		}

		// Retrieve our maas_network_interface_vlan by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		response, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Network interface VLAN (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_network_interface_vlan is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
