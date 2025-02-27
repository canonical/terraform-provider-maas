package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootSources_basic(t *testing.T) {
	os := "ubuntu"
	release := "mantic"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("data.maas_boot_resources.test", "boot_source"),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "os", os),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "release", release),
		resource.TestCheckResourceAttrSet("data.maas_boot_resources.test", "boot_source_selections"),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "boot_source_selections.0.name", fmt.Sprintf("%s/%s", os, release)),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceSelectionDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootReources(os, release),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootReources(os string, release string) string {
	return fmt.Sprintf(`
data "maas_boot_source", "test" {}

data "maas_boot_resources" "test" {
    boot_source = data.maas_boot_source.test.id

	os      = %s
	release = %s
}
`, os, release)
}
