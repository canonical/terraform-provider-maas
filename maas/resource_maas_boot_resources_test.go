package maas_test

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type BootResources struct {
	bootSource           int
	bootSourceSelections []int
}

func TestAccResourceMAASBootResources_basic(t *testing.T) {
	var bootResources BootResources

	checks := []resource.TestCheckFunc{
		testAccMAASBootResourcesCheckExists("maas_boot_resources.test", &bootResources),
		resource.TestCheckResourceAttr("maas_boot_resources.test", "boot_source_selections.#", "1"),
		resource.TestCheckResourceAttrPair("maas_boot_resources.test", "boot_source_selections.0", "maas_boot_source_selection.mantic", "id"),
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

func testAccMAASBootResources() string {
	return `
data "maas_boot_source" "test" {}

resource "maas_boot_source_selection" "mantic" {
  boot_source = data.maas_boot_source.test.id
  os 			= "ubuntu"
  release		= "mantic"
  arches     	= ["*"]
  subarches  	= ["*"]
  labels     	= ["*"]
}

resource "maas_boot_resources" "test" {
  boot_source_selections = [
    maas_boot_source_selection.mantic.id,
  ]
}`
}

func testAccMAASBootResourcesCheckExists(rn string, bootResources *BootResources) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		// fetch the boot source
		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		bootSource, err := conn.BootSources.Get()
		if err != nil {
			return fmt.Errorf("error fetching boot sources: %v", err)
		}

		bootSourceID := bootSource[0].ID

		// fetch the resources
		gotBootResources, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return fmt.Errorf("error fetching synced boot resources: %s", err)
		}

		resourceMap := make(map[string]struct{})
		for _, res := range gotBootResources {
			resourceMap[res.Name] = struct{}{}
		}

		// shenanigans! can't seem to access the selections and then key on top, so we treat it as a single key
		count := rs.Primary.Attributes["boot_source_selections.#"]
		selectionCount, err := strconv.Atoi(count)

		if err != nil {
			return fmt.Errorf("Could not convert %v to integer: %v", count, err)
		}

		if selectionCount < 1 {
			return fmt.Errorf("Boot Resource does not contain any selections!")
		}

		// ensure each selection exists
		var selectionSet []int

		for i := range selectionCount {
			thisID := rs.Primary.Attributes[fmt.Sprintf("boot_source_selections.%d", i)]
			selectionID, err := strconv.Atoi(thisID)

			if err != nil {
				return fmt.Errorf("Could not convert %v to integer: %v", thisID, err)
			}

			selection, err := conn.BootSourceSelection.Get(bootSourceID, selectionID)
			if err != nil {
				return fmt.Errorf("error fetching boot selection %d: %s", selectionID, err)
			}

			if _, exists := resourceMap[fmt.Sprintf("%s/%s", selection.OS, selection.Release)]; !exists {
				return fmt.Errorf("Boot Resource missing for %s/%s", selection.OS, selection.Release)
			}

			selectionSet = append(selectionSet, selection.ID)
		}

		bootResources.bootSource = bootSourceID
		bootResources.bootSourceSelections = selectionSet

		return nil
	}
}

func testAccCheckMAASBootResourcesDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_boot_resources" {
			continue
		}

		// read all of the boot resources
		response, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return fmt.Errorf("error getting synced boot resource: %s", err)
		}

		resourceMap := make(map[string]struct{})
		for _, res := range response {
			resourceMap[res.Name] = struct{}{}
		}

		// fetch the boot source
		bootSource, err := conn.BootSources.Get()
		if err != nil {
			return fmt.Errorf("error fetching boot sources: %v", err)
		}

		bootSourceID := bootSource[0].ID

		// shenanigans to get all the boot selection
		count := rs.Primary.Attributes["boot_source_selections.#"]
		selectionCount, err := strconv.Atoi(count)

		if err != nil {
			return fmt.Errorf("Could not convert %v to integer: %v", count, err)
		}

		if selectionCount < 1 {
			log.Printf("[DEBUG] Boot Resource does not contain any selections, deletion okay")
			return nil
		}

		// ensure each boot selection has been deleted
		for i := range selectionCount {
			thisID := rs.Primary.Attributes[fmt.Sprintf("boot_source_selections.%d", i)]
			selectionID, err := strconv.Atoi(thisID)

			if err != nil {
				return fmt.Errorf("Could not convert %v to integer: %v", thisID, err)
			}

			bootSourceSelection, err := conn.BootSourceSelection.Get(bootSourceID, selectionID)
			if err != nil {
				// 404 means the resource was deleted already
				if strings.Contains(err.Error(), "404 Not Found") {
					return nil
				}

				return fmt.Errorf("error fetching boot selection %d: %s", selectionID, err)
			}

			if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootSourceSelection.OS, bootSourceSelection.Release)]; exists {
				return fmt.Errorf("Boot Resource still exists for %s/%s", bootSourceSelection.OS, bootSourceSelection.Release)
			}
		}

		return nil
	}

	return nil
}

func findBootSourceSelection(client *client.Client, bootSource int, os string, release string) (*entity.BootSourceSelection, error) {
	if bootSourceSelections, err := client.BootSourceSelections.Get(bootSource); err != nil {
		return nil, err
	} else {
		for _, d := range bootSourceSelections {
			if d.OS == os && d.Release == release {
				return &d, nil
			}
		}
	}

	return nil, nil
}
