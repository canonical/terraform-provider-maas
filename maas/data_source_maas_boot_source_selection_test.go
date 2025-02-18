package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootSourceSelection_basic(t *testing.T) {
	os := "ubuntu"
	release := "mantic"
	arches := []string{"amd64"}
	subarches := []string{"*"}
	labels := []string{"*"}

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "boot_source"),
		resource.TestCheckResourceAttr("data.maas_boot_source_selection.test", "os", os),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "release", release),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.0", arches[0]),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "subarches.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "subarches.0", subarches[0]),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "labels.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "labels.0", labels[0]),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceSelectionDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootSourceSelection(os, release, arches, subarches, labels),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootSourceSelection(os string, release string, arches []string, subarches []string, labels []string) string {
	return fmt.Sprintf(`
%s

data "maas_boot_source_selection" "test" {
	boot_source = maas_boot_source_selection.test.boot_source

	os      = maas_boot_source_selection.test.os
	release = maas_boot_source_selection.test.release
}
`, testAccMAASBootSourceSelection(os, release, arches, subarches, labels))
}
