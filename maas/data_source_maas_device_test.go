package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maas/gomaasclient/entity"
)

func TestAccDataSourceMaasDevice_basic(t *testing.T) {

	var device entity.Device
	description := "Test description"
	domain := "test-data-domain"
	hostname := "test-data-device"
	zone := "default"
	mac_address := "12:23:45:67:89:fa"

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
		PreCheck:     func() { testutils.PreCheck(t) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasDeviceDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasDevice(description, domain, hostname, zone, mac_address),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasDevice(description string, domain string, hostname string, zone string, mac_address string) string {
	return fmt.Sprintf(`
%s

data "maas_device" "test" {
	hostname = maas_device.test.hostname
}
`, testAccMaasDevice(description, domain, hostname, zone, mac_address))
}
