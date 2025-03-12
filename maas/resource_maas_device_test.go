package maas_test

import (
	"fmt"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMaasDevice_basic(t *testing.T) {

	var device entity.Device
	description := "Test description"
	domain := acctest.RandomWithPrefix("tf-domain-")
	hostname := acctest.RandomWithPrefix("tf-device-")
	zone := "default"
	mac_address := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMaasDeviceCheckExists("maas_device.test", &device),
		resource.TestCheckResourceAttr("maas_device.test", "description", description),
		resource.TestCheckResourceAttr("maas_device.test", "domain", domain),
		resource.TestCheckResourceAttr("maas_device.test", "fqdn", fmt.Sprintf("%s.%s", hostname, domain)),
		resource.TestCheckResourceAttr("maas_device.test", "hostname", hostname),
		resource.TestCheckResourceAttr("maas_device.test", "zone", zone),
		resource.TestCheckResourceAttr("maas_device.test", "ip_addresses.#", "0"),
		resource.TestCheckResourceAttr("maas_device.test", "network_interfaces.#", "1"),
		resource.TestCheckResourceAttrSet("maas_device.test", "network_interfaces.0.id"),
		resource.TestCheckResourceAttr("maas_device.test", "network_interfaces.0.mac_address", mac_address),
		resource.TestCheckResourceAttr("maas_device.test", "network_interfaces.0.name", "eth0"),
		resource.TestCheckResourceAttrSet("maas_device.test", "owner"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasDeviceDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasDevice(description, domain, hostname, zone, mac_address),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			// Test import using ID
			{
				ResourceName:      "maas_device.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test import using hostname
			{
				ResourceName:      "maas_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_device.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_device.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}
					return rs.Primary.Attributes["hostname"], nil
				},
			},
		},
	})
}

func TestAccResourceMaasDevice_update(t *testing.T) {
	var device entity.Device
	deviceHostname := acctest.RandomWithPrefix("tf-device")
	mac_address := testutils.RandomMAC()
	mac_address2 := testutils.RandomMAC()
	fabricName := acctest.RandomWithPrefix("tf-fabric")
	subnetCIDR := "10.77.77.0/24"
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	subnetGatewayIP := "10.77.77.1"
	linkIPAddress := "10.77.77.42"
	checks := []resource.TestCheckFunc{
		testAccMaasDeviceCheckExists("maas_device.test", &device),
		resource.TestCheckResourceAttr("maas_device.test", "hostname", deviceHostname),
		resource.TestCheckResourceAttr("maas_device.test", "network_interfaces.#", "1"),
		resource.TestCheckResourceAttrSet("maas_device.test", "network_interfaces.0.id"),
		resource.TestCheckResourceAttrSet("maas_device.test", "owner"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasDeviceDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasDeviceNetworkInterfaceConfig(deviceHostname, mac_address, fabricName, subnetCIDR, subnetName, subnetGatewayIP, linkIPAddress),
				Check: resource.ComposeTestCheckFunc(append(
					checks,
					resource.TestCheckResourceAttr("maas_device.test", "network_interfaces.0.mac_address", mac_address),
					resource.TestCheckResourceAttr("maas_device.test", "ip_addresses.#", "0"), // No IP addresses are initially assigned. MAAS will assign them when the device is created.
				)...),
			},
			{
				Config: testAccMaasDeviceNetworkInterfaceConfig(deviceHostname, mac_address2, fabricName, subnetCIDR, subnetName, subnetGatewayIP, linkIPAddress),
				Check:  resource.ComposeTestCheckFunc(append(checks, resource.TestCheckResourceAttr("maas_device.test", "network_interfaces.0.mac_address", mac_address2))...),
			},
		},
	})
}

func testAccMaasDeviceCheckExists(rn string, device *entity.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		gotDevice, err := conn.Device.Get(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting device: %s", err)
		}

		*device = *gotDevice

		return nil
	}
}

func testAccMaasDeviceNetworkInterfaceConfig(deviceHostname string, mac_address string, fabricName string, subnetCIDR string, subnetName string, subnetGatewayIP string, linkIPAddress string) string {
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
	`, deviceHostname, mac_address, fabricName, subnetCIDR, subnetName, subnetGatewayIP, linkIPAddress)
}

func testAccMaasDevice(description string, domain string, hostname string, zone string, mac_address string) string {
	return fmt.Sprintf(`
resource "maas_dns_domain" "test" {
	name          = "%s"
	ttl           = 3600
	authoritative = true
}

resource "maas_device" "test" {
	description        = "%s"
	domain             = maas_dns_domain.test.name
	hostname           = "%s"
	zone               = "%s"
	network_interfaces {
		mac_address = "%s"
	}
}
`, domain, description, hostname, zone, mac_address)
}

func testAccCheckMaasDeviceDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_device
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_device" {
			continue
		}

		// Retrieve our maas_device by referencing it's state ID for API lookup
		response, err := conn.Device.Get(rs.Primary.ID)
		if err == nil {
			if response != nil && response.SystemID == rs.Primary.ID {
				return fmt.Errorf("MAAS Device (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_device is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
