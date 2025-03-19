package maas_test

import (
	"fmt"
	"strconv"
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

func testAccMAASBootResourcesCheckExists(rn string, bootReources *BootResources) resource.TestCheckFunc {
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
		bootsource, err := conn.BootSources.Get()
		if err != nil {
			return fmt.Errorf("error fetching boot sources: %v", err)
		}
		boot_source_id := bootsource[0].ID

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
		for i := 0; i < selectionCount; i++ {
			this_id := rs.Primary.Attributes[fmt.Sprintf("boot_source_selections.%d", i)]
			selection_id, err := strconv.Atoi(this_id)
			if err != nil {
				return fmt.Errorf("Could not convert %v to integer: %v", this_id, err)
			}

			selection, err := conn.BootSourceSelection.Get(boot_source_id, selection_id)
			if err != nil {
				return fmt.Errorf("error fetching boot selection %d: %s", selection_id, err)
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

func testAccCheckMAASBootResourcesDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
	if err := awaitImportComplete(conn); err != nil {
		return fmt.Errorf("Could not await image importing: %v", err)
	}

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
		bootsource, err := conn.BootSources.Get()
		if err != nil {
			return fmt.Errorf("error fetching boot sources: %v", err)
		}
		boot_source_id := bootsource[0].ID

		// shenanigans to get all the boot selection
		count := rs.Primary.Attributes["boot_source_selections.#"]
		selectionCount, err := strconv.Atoi(count)
		if err != nil {
			return fmt.Errorf("Could not convert %v to integer: %v", count, err)
		}
		if selectionCount < 1 {
			return fmt.Errorf("Boot Resource does not contain any selections!")
		}

		// ensure each boot selection has been deleted
		for i := 0; i < selectionCount; i++ {
			this_id := rs.Primary.Attributes[fmt.Sprintf("boot_source_selections.%d", i)]
			selection_id, err := strconv.Atoi(this_id)
			if err != nil {
				return fmt.Errorf("Could not convert %v to integer: %v", this_id, err)
			}
			bootselection, err := conn.BootSourceSelection.Get(boot_source_id, selection_id)
			if err != nil {
				return fmt.Errorf("error fetching boot selection %d: %s", selection_id, err)
			}
			if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootselection.OS, bootselection.Release)]; exists {
				return fmt.Errorf("Boot Resource still exists for %s/%s", bootselection.OS, bootselection.Release)
			}
		}
		return nil
	}

	return nil
}
