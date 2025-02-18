package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASBootSourceSelection_basic(t *testing.T) {

	var bootsourceselection entity.BootSourceSelection
	os := "ubuntu"
	release := "oracular"
	arches := []string{"amd64"}
	subarches := []string{"*"}
	labels := []string{"*"}

	checks := []resource.TestCheckFunc{
		testAccMAASBootSourceSelectionCheckExists("maas_boot_source_selection.test", &bootsourceselection),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "os", os),
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
				Config: testAccMAASBootSourceSelection(os, release, arches, subarches, labels),
				Check:  resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASBootSourceSelectionCheckExists(rn string, bootSourceSelection *entity.BootSourceSelection) resource.TestCheckFunc {
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
		boot_source_id, err := strconv.Atoi(rs.Primary.Attributes["boot_source"])
		if err != nil {
			return err
		}
		gotBootSourceSelection, err := conn.BootSourceSelection.Get(boot_source_id, id)
		if err != nil {
			return fmt.Errorf("error getting boot source selection: %s", err)
		}

		*bootSourceSelection = *gotBootSourceSelection

		return nil
	}
}

func testAccMAASBootSourceSelection(os string, release string, arches []string, subarches []string, labels []string) string {
	return fmt.Sprintf(`
data "maas_boot_source" "test" {}

resource "maas_boot_source_selection" "test" {
	boot_source = data.maas_boot_source.test.id

	os         = "%s"
	release    = "%s"
	arches     = ["%s"]
	subarches  = ["%s"]
	labels     = ["%s"]
}
`, os, release, arches[0], subarches[0], labels[0])
}

func testAccCheckMAASBootSourceSelectionDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_boot_source_selection" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err

		}
		boot_source_id, err := strconv.Atoi(rs.Primary.Attributes["boot_source"])
		if err != nil {
			return err
		}

		response, err := conn.BootSourceSelection.Get(boot_source_id, id)
		if err == nil {
			// default boot source selection leads to noop
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Boot Source Selection (%s %d %d) still exists.", rs.Primary.ID, boot_source_id, id)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_boot_source_selection is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
