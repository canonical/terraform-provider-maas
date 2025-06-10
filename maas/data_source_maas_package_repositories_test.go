package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASPackageRepositories_basic(t *testing.T) {
	resourceName := "main_archive"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "name", resourceName),
		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "url", "http://archive.ubuntu.com/ubuntu"),
		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "disable_sources", "true"),
		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "enabled", "true"),

		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "arches.#", "2"),
		resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "arches.*", "amd64"),
		resource.TestCheckTypeSetElemAttr("maas_package_repository.test_custom", "arches.*", "i386"),

		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "components.#", "0"),

		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "disabled_pockets.#", "0"),

		resource.TestCheckResourceAttr("maas_package_repository.test_custom", "distributions.#", "0"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASPackageRepositories(resourceName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASPackageRepositories(name string) string {
	return fmt.Sprintf(`
data "maas_package_repository" "test" {
  name = %q
}
`, name)
}
