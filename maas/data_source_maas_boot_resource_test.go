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
	arches := []string{"amd64"}
	subarches := []string{"*"}
	labels := []string{"*"}

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("data.maas_boot_resource.test", "boot_source"),
		resource.TestCheckResourceAttrSet("data.maas_boot_resource.test", "boot_source_selections"),
		resource.TestCheckResourceAttr("data.maas_boot_resource.test", "boot_source_selections.#", "1"),
		resource.TestCheckResourceAttr("data.maas_boot_resource.test", "boot_source_selections.0", "data.boot_source_selections.test.id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceSelectionDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootReources(os, release, arches, subarches, labels),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootReources(os string, release string, arches []string, subarches []string, labels []string) string {
	return fmt.Sprintf(`
%s

data "maas_boot_source_selection" "test" {
	boot_source = maas_boot_source_selection.test.boot_source

	os      = maas_boot_source_selection.test.os
	release = maas_boot_source_selection.test.release
}

data "maas_boot_resources" "test" {
    boot_source = data.maas_boot_source.test.id
	
    boot_source_selections = [
        data.maas_boot_source_selection.test.id,
    ]
}
`, testAccMAASBootSourceSelection(os, release, arches, subarches, labels))
}
