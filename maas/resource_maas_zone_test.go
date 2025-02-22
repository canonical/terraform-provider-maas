package maas_test

import (
	"fmt"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceMaasZone_basic(t *testing.T) {
	var zone entity.Zone
	name := acctest.RandomWithPrefix("tf-zone-")
	description := "Test description"

	checks := []resource.TestCheckFunc{
		testAccMaasZoneCheckExists("maas_zone.test", &zone),
		resource.TestCheckResourceAttr("maas_zone.test", "name", name),
		resource.TestCheckResourceAttr("maas_zone.test", "description", description),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasZoneDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasZone(name, description),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			// Test import using name
			{
				ResourceName: "maas_zone.test",
				ImportState:  true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					var zone *terraform.InstanceState
					if len(is) != 1 {
						return fmt.Errorf("expected 1 state: %#v", t)
					}
					zone = is[0]
					assert.Equal(t, zone.Attributes["name"], name)
					assert.Equal(t, zone.Attributes["description"], description)
					return nil
				},
			},
			// Test import using ID
			{
				ResourceName:      "maas_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMaasZoneCheckExists(rn string, zone *entity.Zone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		gotZone, err := getZone(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting zone: %s", err)
		}

		*zone = *gotZone

		return nil
	}
}

func testAccMaasZone(name string, description string) string {
	return fmt.Sprintf(`
resource "maas_zone" "test" {
	name        = "%s"
	description = "%s"
}
`, name, description)
}

func testAccCheckMaasZoneDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_zone
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_zone" {
			continue
		}

		// Retrieve our maas_zone by referencing it's state ID for API lookup
		response, err := getZone(conn, rs.Primary.ID)
		if err == nil {
			if response != nil && fmt.Sprintf("%v", response.ID) == rs.Primary.ID {
				return fmt.Errorf("MAAS Zone (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_zone is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}

func getZone(client *client.Client, identifier string) (*entity.Zone, error) {
	zones, err := client.Zones.Get()
	if err != nil {
		return nil, err
	}
	for _, z := range zones {
		if fmt.Sprintf("%v", z.ID) == identifier || z.Name == identifier {
			return &z, nil
		}
	}
	return nil, fmt.Errorf("404 Not Found: %v", identifier)
}
