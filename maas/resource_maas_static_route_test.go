package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccStaticRoute_basic(t *testing.T) {
	// The source subnet
	sourceCIDR := testutils.GenerateRandomCIDR()
	sourceGateway := testutils.GetNetworkPrefixFromCIDR(sourceCIDR) + ".1"
	sourceName := "source_subnet"

	// The destination subnet
	destinationCIDR := testutils.GenerateRandomCIDR()
	destinationGateway := testutils.GetNetworkPrefixFromCIDR(destinationCIDR) + ".1"
	destinationName := "destination_subnet"

	// And a third subnet we will change the static route to point to instead
	changedCIDR := testutils.GenerateRandomCIDR()
	changedGateway := testutils.GetNetworkPrefixFromCIDR(changedCIDR) + ".1"
	changedName := "changed_subnet"

	// we create a base configuration containing all the subnets that will be used in the tests
	baseConfig := testAccGenerateSubnet(sourceCIDR, sourceGateway, sourceName) +
		testAccGenerateSubnet(destinationCIDR, destinationGateway, destinationName) +
		testAccGenerateSubnet(changedCIDR, changedGateway, changedName)

	metric := 55
	newMetric := 40

	// the actual test construct go test uses
	resource.ParallelTest(t, resource.TestCase{
		// boilerplate every basic test has, we provide basic pre-check and error check
		// as our test doesn't require anything fancy
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		// When the test concludes we'll need to test the resources are correctly destroyed
		CheckDestroy: testAccCheckMAASStaticRouteDestroy,
		// We define the tests that will be run against our resource
		Steps: []resource.TestStep{
			// Test the resource can be created
			{
				// The Terraform plan to create the resource
				Config: baseConfig + testAccStaticRouteConfig(sourceName, destinationName, metric),
				// And the checks that should be run to determine if the resource was created correctly
				Check: resource.ComposeTestCheckFunc(
					// first check the resource exists
					testAccCheckStaticRouteExists("maas_static_route.test"),
					// and then that it has the expected properties
					resource.TestCheckResourceAttr("maas_static_route.test", "gateway_ip", sourceGateway),
					resource.TestCheckResourceAttr("maas_static_route.test", "metric", fmt.Sprintf("%d", metric)),
					resource.TestCheckResourceAttr("maas_static_route.test", "source", sourceName),
					resource.TestCheckResourceAttr("maas_static_route.test", "destination", destinationName),
				),
			},
			// Test the resource can be updated
			{
				// The Terraform plan to create the resource
				Config: baseConfig + testAccStaticRouteConfig(sourceName, changedName, newMetric),
				// And the checks that should be run to determine if the resource was created correctly
				Check: resource.ComposeTestCheckFunc(
					// first check the resource exists
					testAccCheckStaticRouteExists("maas_static_route.test"),
					// and then that it has the expected properties
					resource.TestCheckResourceAttr("maas_static_route.test", "gateway_ip", sourceGateway),
					resource.TestCheckResourceAttr("maas_static_route.test", "metric", fmt.Sprintf("%d", newMetric)),
					resource.TestCheckResourceAttr("maas_static_route.test", "source", sourceName),
					resource.TestCheckResourceAttr("maas_static_route.test", "destination", changedName),
				),
			},
			// And finally, we need to test the import functionality
			{
				ResourceName:      "maas_static_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// a helper function to generate subnets
func testAccGenerateSubnet(cidr string, gateway string, name string) string {
	return fmt.Sprintf(`
resource "maas_subnet" "test_subnet_%s" {
  cidr       = %q
  gateway_ip = %q
  fabric     = 0
  vlan       = 0
  name       = %q

  dns_servers = [
    "1.1.1.1",
  ]
}
`, name, cidr, gateway, name)
}

// We generate the references from the source and destination names, ensuring the gateway is that of the source
func testAccStaticRouteConfig(source string, destination string, metric int) string {
	return fmt.Sprintf(`
resource "maas_static_route" "test" {
  source      = maas_subnet.test_subnet_%s.name
  destination = maas_subnet.test_subnet_%s.name
  metric      = %d
  gateway_ip  = maas_subnet.test_subnet_%s.gateway_ip
}`, source, destination, metric, source)
}

// Check whether the resource exists at MAAS.
func testAccCheckStaticRouteExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", resourceName, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		if _, err = conn.StaticRoute.Get(id); err != nil {
			return fmt.Errorf("error getting the Static Route: %s", err)
		}

		return nil
	}
}

// Check whether the resource has been deleted at MAAS
func testAccCheckMAASStaticRouteDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// we loop through all resources in the state
	for _, rs := range s.RootModule().Resources {
		// ignoring anything that isn't a static route
		if rs.Type != "maas_static_route" {
			continue
		}

		// To determine if the route has been destroyed, we attempt to GET
		// it from the MAAS API
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := conn.StaticRoute.Get(id)
		if err == nil {
			// if it succeeded, the route must still exist
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Static Route (%s) still exists.", rs.Primary.ID)
			}
		}

		// error 404 means the route wasn't found, and the resource is deleted
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
