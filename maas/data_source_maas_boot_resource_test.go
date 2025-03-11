package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootResources_basic(t *testing.T) {
	os := "ubuntu"
	release := "kinetic"
	arches := []string{"*"}
	subarches := []string{"*"}
	labels := []string{"*"}

	checks := []resource.TestCheckFunc{
		// We check the selection was imported correctly
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "os", os),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "release", release),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.0", arches[0]),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "subarches.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "subarches.0", subarches[0]),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "labels.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "labels.0", labels[0]),
		// and then that the resources are populated correctly too
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "os", os),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "release", release),
		resource.TestCheckResourceAttrSet("data.maas_boot_resources.test", "boot_resources.#"),
		resource.TestCheckResourceAttr("data.maas_boot_resources.test", "boot_resources.0.name", fmt.Sprintf("%s/%s", os, release)),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootResourcesDestroy,
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

data "maas_boot_resources" "test" {
	os      = maas_boot_source_selection.test.os
	release = maas_boot_source_selection.test.release
}
`, testAccMAASBootSourceSelection(os, release, arches, subarches, labels))
}
