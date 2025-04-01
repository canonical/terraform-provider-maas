package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMAASDevice_basic(t *testing.T) {
	var device entity.Device

	description := "Test description"
	domain := acctest.RandomWithPrefix("tf-domain-")
	hostname := acctest.RandomWithPrefix("tf-device-")
	zone := "default"
	macAddress := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMAASDeviceCheckExists("maas_device.test", &device),
		resource.TestCheckResourceAttr("data.maas_device.test", "description", description),
		resource.TestCheckResourceAttr("data.maas_device.test", "domain", domain),
		resource.TestCheckResourceAttr("data.maas_device.test", "fqdn", fmt.Sprintf("%s.%s", hostname, domain)),
		resource.TestCheckResourceAttr("data.maas_device.test", "hostname", hostname),
		resource.TestCheckResourceAttr("data.maas_device.test", "zone", zone),
		resource.TestCheckResourceAttr("data.maas_device.test", "ip_addresses.#", "0"),
		resource.TestCheckResourceAttr("data.maas_device.test", "network_interfaces.#", "1"),
		resource.TestCheckResourceAttrSet("data.maas_device.test", "network_interfaces.0.id"),
		resource.TestCheckResourceAttr("data.maas_device.test", "network_interfaces.0.mac_address", macAddress),
		resource.TestCheckResourceAttr("data.maas_device.test", "network_interfaces.0.name", "eth0"),
		resource.TestCheckResourceAttrSet("data.maas_device.test", "owner"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASDeviceDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASDevice(description, domain, hostname, zone, macAddress),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASDevice(description string, domain string, hostname string, zone string, macAddress string) string {
	return fmt.Sprintf(`
%s

data "maas_device" "test" {
	hostname = maas_device.test.hostname
}
`, testAccMAASDevice(description, domain, hostname, zone, macAddress))
}
