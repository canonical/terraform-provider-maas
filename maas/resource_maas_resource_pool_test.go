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

func TestAccResourceMaasResourcePool_basic(t *testing.T) {

	var resourcePool entity.ResourcePool
	description := "Test description"
	name := acctest.RandomWithPrefix("tf-resource-pool-")

	checks := []resource.TestCheckFunc{
		testAccMaasResourcePoolCheckExists("maas_resource_pool.test", &resourcePool),
		resource.TestCheckResourceAttr("maas_resource_pool.test", "description", description),
		resource.TestCheckResourceAttr("maas_resource_pool.test", "name", name),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasResourcePoolDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasResourcePool(description, name),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			// Test import using ID
			{
				ResourceName:      "maas_resource_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test import using name
			{
				ResourceName:      "maas_resource_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_resource_pool.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_resource_pool.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}
					return rs.Primary.Attributes["name"], nil
				},
			},
		},
	})
}

func testAccMaasResourcePoolCheckExists(rn string, resourcePool *entity.ResourcePool) resource.TestCheckFunc {
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
		gotResourcePool, err := conn.ResourcePool.Get(id)
		if err != nil {
			return fmt.Errorf("error getting resource pool: %s", err)
		}

		*resourcePool = *gotResourcePool

		return nil
	}
}

func testAccMaasResourcePool(description string, name string) string {
	return fmt.Sprintf(`
resource "maas_resource_pool" "test" {
	name        = "%s"
	description = "%s"
}
`, name, description)
}

func testAccCheckMaasResourcePoolDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_resource_pool
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_resource_pool" {
			continue
		}

		// Retrieve our maas_resource_pool by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		response, err := conn.ResourcePool.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Resource pool (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_resource_pool is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
