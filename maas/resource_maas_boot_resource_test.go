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

type BootResources struct {
	boot_source            int
	boot_source_selections []int
}

func TestAccResourceMAASBootResources_basic(t *testing.T) {

	var bootresources BootResources

	checks := []resource.TestCheckFunc{
		testAccMAASBootResourcesCheckExists("maas_boot_resources.test", &bootresources),
		resource.TestCheckResourceAttr("maas_boot_resources.test", "boot_source_selections.#", "3"),
		resource.TestCheckResourceAttr("maas_boot_resources.test", "boot_source_selections.0.os", "ubuntu"),
		resource.TestCheckResourceAttr("maas_boot_resources.test", "boot_source_selections.0.release", "mantic"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootResourcesDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASBootResources(),
				Check:  resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASBootResourcesCheckExists(rn string, bootReources *BootResources) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		bootsource, err := conn.BootSources.Get()
		if err != nil {
			return err
		}
		boot_source_id := bootsource[0].ID

		gotBootResources, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return fmt.Errorf("error getting boot resource: %s", err)
		}
		resourceMap := make(map[string]struct{})
		for _, res := range gotBootResources {
			resourceMap[res.Name] = struct{}{}
		}

		var selectionSet []int
		for _, sel := range strings.Split(rs.Primary.Attributes["boot_source_selection"], ",") {
			selection_id, err := strconv.Atoi(sel)
			if err != nil {
				return err
			}

			selection, err := conn.BootSourceSelection.Get(boot_source_id, selection_id)
			if err != nil {
				return err
			}
			if _, exists := resourceMap[fmt.Sprintf("%s/%s", selection.OS, selection.Release)]; !exists {
				return fmt.Errorf("Boot Resource missing for %s/%s", selection.OS, selection.Release)
			}
			selectionSet = append(selectionSet, selection.ID)
		}

		bootReources.boot_source = boot_source_id
		bootReources.boot_source_selections = selectionSet

		return nil
	}
}

func testAccMAASBootResources() string {
	return `
data "maas_boot_source" "test" {}

resource "maas_boot_source_selection" "mantic" {
    boot_source = data.maas_boot_source.test.id
    os = "ubuntu"
    release = "mantic"
}

resource "maas_boot_resources" "test" {
    boot_source_selections = [
        maas_boot_source_selection.mantic.id,
    ]
}`
}

func testAccCheckMAASBootResourcesDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_boot_resources" {
			continue
		}

		response, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return err
		}
		resourceMap := make(map[string]struct{})
		for _, res := range response {
			resourceMap[res.Name] = struct{}{}
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		bootsource, err := conn.BootSources.Get()
		if err != nil {
			return err
		}
		boot_source_id := bootsource[0].ID

		// an empty string means no selections present
		existing_selections := rs.Primary.Attributes["boot_source_selections"]
		if existing_selections == "" {
			continue
		}

		for _, selection := range strings.Split(existing_selections, ",") {
			selection_id, err := strconv.Atoi(selection)
			if err != nil {
				return err
			}
			bootselection, err := conn.BootSourceSelection.Get(boot_source_id, selection_id)
			if err != nil {
				return err
			}
			if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootselection.OS, bootselection.Release)]; exists {
				return fmt.Errorf("Boot Resource still exists for %s/%s", bootselection.OS, bootselection.Release)
			}
		}
		return nil
	}

	return nil
}
