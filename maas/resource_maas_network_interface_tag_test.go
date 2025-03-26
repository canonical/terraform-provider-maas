package maas_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestSplitTagStateId(t *testing.T) {
	expectedSystemId := "abc123"
	expectedInterfaceId := 12
	stateId := fmt.Sprintf("%s/%d", expectedSystemId, expectedInterfaceId)
	systemId, interfaceId, err := maas.SplitTagStateId(stateId)
	if err != nil {
		t.Fatalf("Error splitting state ID: %s", err)
	}
	if systemId != expectedSystemId || interfaceId != expectedInterfaceId {
		t.Fatalf("Expected system ID %s and interface ID %d, got system ID %s and interface ID %d", expectedSystemId, expectedInterfaceId, systemId, interfaceId)
	}
}

func TestAccNetworkInterfaceTag_basic(t *testing.T) {
	hostname := acctest.RandomWithPrefix("tf")
	macAddress := testutils.RandomMAC()
	tagName := acctest.RandomWithPrefix("tag")
	tagName2 := acctest.RandomWithPrefix("tag")
	tagName3 := acctest.RandomWithPrefix("tag")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMaasNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			// Test creation.
			{
				Config: testAccMaasNetworkInterfaceTagConfig(hostname, macAddress, tagName, tagName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaasNetworkInterfaceTagExists("maas_network_interface_tag.test", tagName, tagName2),
					resource.TestCheckResourceAttr("maas_network_interface_tag.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("maas_network_interface_tag.test", "tags.*", tagName),
					resource.TestCheckTypeSetElemAttr("maas_network_interface_tag.test", "tags.*", tagName2),
				),
			},
			// Test update. Expected behaviour is that the previous tag is removed and the new tag is added.
			{
				Config: testAccMaasNetworkInterfaceTagConfig(hostname, macAddress, tagName2, tagName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaasNetworkInterfaceTagExists("maas_network_interface_tag.test", tagName2, tagName3),
					resource.TestCheckResourceAttr("maas_network_interface_tag.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("maas_network_interface_tag.test", "tags.*", tagName2),
					resource.TestCheckTypeSetElemAttr("maas_network_interface_tag.test", "tags.*", tagName3),
				),
			},
			// Test import.
			{
				ResourceName:      "maas_network_interface_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_network_interface_tag.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_network_interface_tag.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["device"], rs.Primary.Attributes["interface_id"]), nil
				},
			},
		},
	})
}

func testAccMaasNetworkInterfaceTagConfig(hostname, macAddress string, tagNames ...string) string {
	return fmt.Sprintf(`
resource "maas_device" "test" {
  hostname = %q
  network_interfaces {
    mac_address = %q
  }
}

resource "maas_network_interface_tag" "test" {
  device       = maas_device.test.id
  interface_id = [for iface in maas_device.test.network_interfaces : iface.id if iface.mac_address == %q][0]
  tags         = %s
}
	`, hostname, macAddress, macAddress, fmt.Sprintf("[\"%s\"]", strings.Join(tagNames, "\", \"")))
}

func testAccCheckMaasNetworkInterfaceTagExists(resourceName string, tagNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		// Get the system and interface ID from the state ID
		systemId, interfaceId, err := maas.SplitTagStateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Get the existing interface
		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		response, err := conn.NetworkInterface.Get(systemId, interfaceId)
		if err != nil {
			return err
		}
		if response == nil {
			return fmt.Errorf("MAAS Network Interface (%s) not found.", rs.Primary.ID)
		}

		// Check the tags exist
		for _, tag := range tagNames {
			if !slices.Contains(response.Tags, tag) {
				return fmt.Errorf("MAAS Network Interface (%s) tags (%s) do not match expected tags: %s", rs.Primary.ID, response.Tags, tagNames)
			}
		}

		return nil
	}
}

func testAccCheckMaasNetworkInterfaceDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_network_interface_tag is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_network_interface_tag" {
			continue
		}

		// Retrieve the system and interface ID from the state ID
		systemId, interfaceId, err := maas.SplitTagStateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Check the interface doesn't exist
		response, err := conn.NetworkInterface.Get(systemId, interfaceId)
		if err == nil {
			if response != nil && response.ID == interfaceId {
				return fmt.Errorf("MAAS Network Interface (%s) still exists.", rs.Primary.ID)
			}
		}
		// If the error is not a 404, the interface has not been destroyed as it should have been
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}
	return nil
}
