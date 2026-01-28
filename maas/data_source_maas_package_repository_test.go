package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASPackageRepository_basic(t *testing.T) {
	resourceName := "main_archive"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "name", resourceName),
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "url", "http://archive.ubuntu.com/ubuntu"),
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "disable_sources", "true"),
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "enabled", "true"),

		resource.TestCheckResourceAttr("data.maas_package_repository.test", "arches.#", "2"),
		resource.TestCheckTypeSetElemAttr("data.maas_package_repository.test", "arches.*", "amd64"),
		resource.TestCheckTypeSetElemAttr("data.maas_package_repository.test", "arches.*", "i386"),
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "components.#", "0"),
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "disabled_pockets.#", "0"),
		resource.TestCheckResourceAttr("data.maas_package_repository.test", "distributions.#", "0"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASPackageRepository(resourceName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASPackageRepository(name string) string {
	return fmt.Sprintf(`
data "maas_package_repository" "test" {
  name = %q
}
`, name)
}
