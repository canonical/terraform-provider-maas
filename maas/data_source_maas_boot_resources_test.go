package maas_test

import (
	"fmt"
	"os"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASBootResources_basic(t *testing.T) {
	resourcesOS := "ubuntu"
	resourcesRelease := os.Getenv("TF_ACC_BOOT_RESOURCES_OS")

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "os", resourcesOS),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "release", resourcesRelease),
		resource.TestCheckResourceAttrSet("data.maas_boot_resources.test", "boot_resources.#"),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "boot_resources.0.name", fmt.Sprintf("%s/%s", resourcesOS, resourcesRelease)),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_BOOT_RESOURCES_OS"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASBootResources(resourcesRelease),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASBootResources(release string) string {
	return fmt.Sprintf(`
data "maas_boot_resources" "test" {
  os      = "ubuntu"
  release = %q
}
`, release)
}
