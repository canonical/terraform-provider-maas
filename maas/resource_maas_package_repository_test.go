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
	customChangedRepo := testAccCustomPackageRepository(
		"test_custom",
		"custom changed repo",
		"secretKey2",
		"https://test2.com",
		true,
		true,
		[]string{"armhf"},
		[]string{"restricted"},
		[]string{"security"},
		[]string{"jammy-prod"},
	)

	ubuntuRepo := testAccUbuntuPackageRepository(
		"test_ubuntu",
		"main_archive",
		"",
		"http://archive.ubuntu.com/ubuntu",
		false,
		true,
		[]string{"amd64", "i386"},
		[]string{},
		[]string{},
		[]string{},
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckPackageRepositoryDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test creation of custom repo
			{
				Config: customRepo,
				Check: resource.ComposeTestCheckFunc(
					testAccPackageRepositoryCheckExists("maas_package_repository.test_custom"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "name", "custom repo"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "url", "https://test.com"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "disable_sources", "true"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "enabled", "true"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "arches.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "arches.*", "amd64"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "components.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "components.*", "main"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "disabled_pockets.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "disabled_pockets.*", "updates"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "distributions.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "distributions.*", "jammy-prod"),
				),
			},
			// Test updates of custom repo
			{
				Config: customChangedRepo,
				Check: resource.ComposeTestCheckFunc(
					testAccPackageRepositoryCheckExists("maas_package_repository.test_custom"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "name", "custom changed repo"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "url", "https://test2.com"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "disable_sources", "true"),
					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "enabled", "true"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "arches.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "arches.*", "armhf"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "components.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "components.*", "restricted"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "disabled_pockets.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "disabled_pockets.*", "security"),

					resource.TestCheckResourceAttr("maas_package_repository.test_custom", "distributions.#", "1"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "distributions.*", "jammy-prod"),
				),
			},
			// Test import using ID
			{
				Config:        ubuntuRepo,
				ResourceName:  "maas_package_repository.test_ubuntu",
				ImportState:   true,
				ImportStateId: "1",
			},
			// Test importing with Name, only makes sense for custom repos.
			{
				ResourceName:  "maas_package_repository.test_custom",
				ImportState:   true,
				ImportStateId: "custom changed repo",
			},
			// Test importing with URL
			{
				Config:        ubuntuRepo,
				ResourceName:  "maas_package_repository.test_ubuntu",
				ImportState:   true,
				ImportStateId: "http://archive.ubuntu.com/ubuntu",
			},
		},
	},
	)
}

func TestAccResourceMAASPackageRepository_validation(t *testing.T) {
	ubuntuSecurityRepo := testAccUbuntuPackageRepository(
		"test_ubuntu",
		"security.ubuntu.com",
		"",
		"http://security.ubuntu.com/ubuntu",
		false,
		false,
		[]string{},
		[]string{},
		[]string{},
		[]string{},
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckPackageRepositoryDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test security repo, disabled, with no architectures listed
			{
				Config: ubuntuSecurityRepo,
				Check: resource.ComposeTestCheckFunc(
					testAccPackageRepositoryCheckExists("maas_package_repository.test_ubuntu"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "name", "security.ubuntu.com"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "url", "http://security.ubuntu.com/ubuntu"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "disable_sources", "false"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "enabled", "false"),

					// MAAS defaults to amd64 and i386 architectures if none specified
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "arches.#", "2"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_ubuntu", "arches.*", "amd64"),
					resource.TestCheckTypeSetElemAttr("maas_package_repository.test_ubuntu", "arches.*", "i386"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "components.#", "0"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "disabled_pockets.#", "0"),
					resource.TestCheckResourceAttr("maas_package_repository.test_ubuntu", "distributions.#", "0"),
				),
			},
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
	resource := fmt.Sprintf(`
resource "maas_package_repository" %q {
  name = %q
  key  = %q
  url  = %q

  disable_sources = %t
  enabled         = %t

  disabled_components = %v
  disabled_pockets = %v
  distributions = %v
`, resourceName, name, key, url, disableSources, enabled, listAsString(disabledComponents), listAsString(disabledPockets), listAsString(distributions))

	if len(arches) > 0 {
		resource += fmt.Sprintf(`
  arches = %v
`, listAsString(arches))
	}

	return resource + "}\n"
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
		return "[]"
	}

	asList, _ := json.Marshal(stringList)

	return string(asList)
}
