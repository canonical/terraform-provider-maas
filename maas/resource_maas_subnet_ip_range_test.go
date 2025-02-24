package maas_test

import (
	"fmt"
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

func TestAccResourceMaasSubnetIPRange_basic(t *testing.T) {
	// Setup IP range parameters
	var ipRange entity.IPRange
	subnet_name := acctest.RandomWithPrefix("test-subnet")
	ipRangeAttrName := "maas_subnet_ip_range.test_ip_range"
	range_type := "reserved"
	comment := "test-comment"
	ipStart := "10.88.88.1"
	ipEnd := "10.88.88.50"
	ipStartMod := "10.88.88.2"
	ipEndMod := "10.88.88.49"
	commentMod := "a-different-comment"

	// Check functions
	checks := []resource.TestCheckFunc{
		testAccMAASSubnetIPRangeCheckExists(ipRangeAttrName, &ipRange),
		resource.TestCheckResourceAttr(ipRangeAttrName, "type", range_type),
		resource.TestCheckResourceAttr(ipRangeAttrName, "comment", comment),
		resource.TestCheckResourceAttr(ipRangeAttrName, "start_ip", ipStart),
		resource.TestCheckResourceAttr(ipRangeAttrName, "end_ip", ipEnd),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASSubnetIPRangeDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccSubnetIPRangeExampleResource(
					subnet_name,
					range_type,
					comment,
					ipStart,
					ipEnd,
				),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
			// Test if resource drift is detected by modifying the IP range using the
			// Go MAAS client, and then checking if the state is updated correctly on
			// refresh
			{
				PreConfig: func() {
					client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
					params := entity.IPRangeParams{
						Type:    range_type,
						Comment: commentMod,
						StartIP: ipStartMod,
						EndIP:   ipEndMod,
						Subnet:  strconv.Itoa(ipRange.Subnet.ID),
					}
					// Update the IP range to changed values that should be read into state
					_, err := client.IPRange.Update(ipRange.ID, &params)
					if err != nil {
						t.Fatalf("Failed to update IP range: %s", err)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ipRangeAttrName, "start_ip", ipStartMod),
					resource.TestCheckResourceAttr(ipRangeAttrName, "end_ip", ipEndMod),
					resource.TestCheckResourceAttr(ipRangeAttrName, "comment", commentMod),
					resource.TestCheckResourceAttr(ipRangeAttrName, "type", range_type),
				),
			},
			// Test import
			{
				ResourceName:      ipRangeAttrName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test import using start_ip:end_ip format
			{
				ResourceName:      ipRangeAttrName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s:%s", ipStartMod, ipEndMod),
			},
		},
	})
}

// Check if the IP range specified actually exists in MAAS
func testAccMAASSubnetIPRangeCheckExists(rn string, ipRange *entity.IPRange) resource.TestCheckFunc {
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
		gotIPRange, err := conn.IPRange.Get(id)
		if err != nil {
			return fmt.Errorf("error getting ip range: %s", err)
		}

		*ipRange = *gotIPRange

		return nil
	}
}

// An example resource configuration for a subnet IP range.
// Note that a subnet is required to create an IP range.
func testAccSubnetIPRangeExampleResource(
	name_subnet string,
	range_type string,
	comment string,
	start_ip string,
	end_ip string,
) string {
	// A subnet is required to create an IP range
	return fmt.Sprintf(`
		resource "maas_subnet" "test_subnet" {
		  cidr        = "10.88.88.0/26"
		  name        = "%s"
		  gateway_ip  = "10.88.88.1"
		  dns_servers = ["8.8.8.8"]
		}

		resource "maas_subnet_ip_range" "test_ip_range" {
		  subnet   = maas_subnet.test_subnet.id
		  type     = "%s"
		  start_ip = "%s"
		  end_ip   = "%s"
		  comment  = "%s"
		}
	`, name_subnet, range_type, start_ip, end_ip, comment)
}

func testAccCheckMAASSubnetIPRangeDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_subnet_ip_range
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_subnet_ip_range" {
			continue
		}

		// retrieve the maas_subnet_ip_range by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Check if the IP range exists
		var exists bool

		response, err := conn.IPRange.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				exists = true
			}
		}

		if exists {
			return fmt.Errorf("MAAS %s (%s) still exists.", rs.Type, rs.Primary.ID)
		}

		// If the error is equivalent to 404 not found, the maas_resource_pool is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
