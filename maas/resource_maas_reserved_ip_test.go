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

func TestAccResourceMAASReservedIP_basic(t *testing.T) {
	testutils.SkipTestIfNotMAASVersion(t, ">=3.6.0")

	cidr := testutils.GenerateRandomCIDR()
	subnetName := acctest.RandomWithPrefix("tf-reserved-ip-test")
	ip := testutils.GetNetworkPrefixFromCIDR(cidr) + ".50"
	macAddress := testutils.RandomMAC()
	comment := "test static lease"
	commentMod := "updated comment"
	attrName := "maas_reserved_ip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASReservedIPDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test create
			{
				Config: testAccReservedIPConfig(cidr, subnetName, ip, macAddress, comment),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(attrName, "ip", ip),
					resource.TestCheckResourceAttr(attrName, "mac_address", macAddress),
					resource.TestCheckResourceAttr(attrName, "comment", comment),
					resource.TestCheckResourceAttrSet(attrName, "subnet"),
				),
			},
			// Test update
			{
				Config: testAccReservedIPConfig(cidr, subnetName, ip, macAddress, commentMod),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(attrName, "comment", commentMod),
				),
			},
			// Test import
			{
				ResourceName:      attrName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test explicit destroy: remove reserved IP from config while keeping subnet.
			// Otherwise check destroy will always pass because subnet deletion will remove
			// the reserved IP as well.
			{
				Config: testAccSubnetOnlyConfig(cidr, subnetName),
				Check:  testAccCheckMAASReservedIPDeleted(ip),
			},
		},
	})
}

func testAccCheckMAASReservedIPDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_reserved_ip" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := conn.ReservedIP.Get(id)

		if err == nil && response != nil && response.ID == id {
			return fmt.Errorf("MAAS %s (%s) still exists.", rs.Type, rs.Primary.ID)
		}

		if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}

func testAccSubnetOnlyConfig(cidr, subnetName string) string {
	return fmt.Sprintf(`
resource "maas_subnet" "test" {
  cidr = %q
  name = %q
}
`, cidr, subnetName)
}

func testAccCheckMAASReservedIPDeleted(ip string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		rips, err := conn.ReservedIPs.Get()
		if err != nil {
			return fmt.Errorf("error listing reserved IPs: %s", err)
		}

		for _, rip := range rips {
			if rip.IP == ip {
				return fmt.Errorf("reserved IP %s still exists in MAAS", ip)
			}
		}

		return nil
	}
}

func testAccReservedIPConfig(cidr, subnetName, ip, macAddress, comment string) string {
	return testAccSubnetOnlyConfig(cidr, subnetName) + fmt.Sprintf(`
resource "maas_reserved_ip" "test" {
  ip          = %q
  mac_address = %q
  subnet      = maas_subnet.test.id
  comment     = %q
}
`, ip, macAddress, comment)
}

func TestAccResourceMAASReservedIP_noSubnetField(t *testing.T) {
	testutils.SkipTestIfNotMAASVersion(t, ">=3.6.0")

	cidr := testutils.GenerateRandomCIDR()
	subnetName := acctest.RandomWithPrefix("tf-reserved-ip-test")
	ip := testutils.GetNetworkPrefixFromCIDR(cidr) + ".51"
	macAddress := testutils.RandomMAC()
	comment := "test auto-detect subnet"
	attrName := "maas_reserved_ip.test_no_subnet"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASReservedIPDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Create without explicit subnet
			{
				Config: testAccReservedIPConfigNoSubnet(cidr, subnetName, ip, macAddress, comment),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(attrName, "ip", ip),
					resource.TestCheckResourceAttr(attrName, "mac_address", macAddress),
					resource.TestCheckResourceAttr(attrName, "comment", comment),
					resource.TestCheckResourceAttrSet(attrName, "subnet"),
				),
			},
			// Test import
			{
				ResourceName:      attrName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test destroy
			{
				Config: testAccSubnetOnlyConfig(cidr, subnetName),
				Check:  testAccCheckMAASReservedIPDeleted(ip),
			},
		},
	})
}

func testAccReservedIPConfigNoSubnet(cidr, subnetName, ip, macAddress, comment string) string {
	return testAccSubnetOnlyConfig(cidr, subnetName) + fmt.Sprintf(`
resource "maas_reserved_ip" "test_no_subnet" {
  ip          = %q
  mac_address = %q
  comment     = %q

  depends_on = [maas_subnet.test]
}
`, ip, macAddress, comment)
}
