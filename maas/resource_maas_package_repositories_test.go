package maas_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASPackageRepository_basic(t *testing.T) {
	customRepo := testAccCustomPackageRepository(
		"test_custom",
		"custom repo",
		"secretKey",
		"https://test.com",
		true,
		true,
		[]string{"amd64"},
		[]string{"main"},
		[]string{"updates"},
		[]string{"jammy-prod"},
	)

	ubuntuRepo := testAccUbuntuPackageRepository(
		"test_ubuntu",
		"ubuntu repo",
		"secretKey",
		"http://ports.ubuntu.com/",
		true,
		true,
		[]string{"amd64"},
		[]string{"universe"},
		[]string{"updates"},
		[]string{"jammy-prod"},
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckPackageRepositoryDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: customRepo + ubuntuRepo,
				Check: resource.ComposeTestCheckFunc(
					testAccPackageRepositoryCheckExists("maas_package_repository.test_custom"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "name", "custom repo"),
				),
			},
			// Test updates
		},
	},
	)
}

func testAccCustomPackageRepository(resourceName string, name string, key string, url string, disableSources bool, enabled bool, arches []string, components []string, disabledPockets []string, distributions []string) string {
	return fmt.Sprintf(`
resource "maas_package_repository" %q {
  name = %q
  key  = %q
  url  = %q

  disable_sources = %t
  enabled         = %t

  arches = %v
  components = %v
  disabled_pockets = %v
  distributions = %v
}
`, resourceName, name, key, url, disableSources, enabled, listAsString(arches), listAsString(components), listAsString(disabledPockets), listAsString(distributions))
}

func testAccUbuntuPackageRepository(resourceName string, name string, key string, url string, disableSources bool, enabled bool, arches []string, disabledComponents []string, disabledPockets []string, distributions []string) string {
	return fmt.Sprintf(`
resource "maas_package_repository" %q {
  name = %q
  key  = %q
  url  = %q

  disable_sources = %t
  enabled         = %t

  arches = %v
  disabled_components = %v
  disabled_pockets = %v
  distributions = %v
}
`, resourceName, name, key, url, disableSources, enabled, listAsString(arches), listAsString(disabledComponents), listAsString(disabledPockets), listAsString(distributions))
}

func testAccPackageRepositoryCheckExists(rn string) resource.TestCheckFunc {
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

		_, err = conn.PackageRepository.Get(id)
		if err != nil {
			return fmt.Errorf("error getting package repository: %s", err)
		}

		return nil
	}
}

func testAccCheckPackageRepositoryDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_package_repository" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := conn.PackageRepository.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("Package Repository %s (%d) still exists.", response.Name, id)
			}
		}

		// 404 means destroyed, anything else is an error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}

func listAsString(stringList []string) string {
	if len(stringList) == 0 {
		return ""
	}

	asList, _ := json.Marshal(stringList)

	return string(asList)
}
