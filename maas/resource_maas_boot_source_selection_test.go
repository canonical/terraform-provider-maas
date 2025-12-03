package maas_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/go-set/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASBootSourceSelection_basic(t *testing.T) {
	var bootsourceselection entity.BootSourceSelection

	os := "ubuntu"
	release := "oracular"
	arches := []string{"ppc64el"}
	updatedArches := []string{"arm64", "s390x"}

	checks := []resource.TestCheckFunc{
		testAccMAASBootSourceSelectionCheckExists("maas_boot_source_selection.test", &bootsourceselection),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "os", os),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "release", release),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "subarches.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "subarches.0", "*"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "labels.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "labels.0", "*"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceSelectionDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccMAASBootSourceSelection(os, release, arches),
				Check: resource.ComposeAggregateTestCheckFunc(append(
					checks,
					resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.#", fmt.Sprintf("%v", len(arches))),
					resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.0", arches[0]),
				)...,
				),
			},
			// Test update
			{
				Config: testAccMAASBootSourceSelection(os, release, updatedArches),
				Check: resource.ComposeAggregateTestCheckFunc(append(
					checks,
					resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.#", fmt.Sprintf("%v", len(updatedArches))),
					resource.TestCheckTypeSetElemAttr("maas_boot_source_selection.test", "arches.*", updatedArches[0]),
					resource.TestCheckTypeSetElemAttr("maas_boot_source_selection.test", "arches.*", updatedArches[1]),
				)...,
				),
			},
			// Test import
			{
				ResourceName:      "maas_boot_source_selection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_boot_source_selection.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_boot_source_selection.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}

					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["boot_source"], rs.Primary.ID), nil
				},
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

		bootSourceID, err := strconv.Atoi(rs.Primary.Attributes["boot_source"])
		if err != nil {
			return err
		}

		gotBootSourceSelection, err := conn.BootSourceSelection.Get(bootSourceID, id)
		if err != nil {
			return fmt.Errorf("error getting boot source selection: %s", err)
		}

		arches := []string{}

		archesLen, err := strconv.Atoi(rs.Primary.Attributes["arches.#"])
		if err != nil {
			return err
		}

		for i := range archesLen {
			arches = append(arches, rs.Primary.Attributes[fmt.Sprintf("arches.%v", i)])
		}

		allResources, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return err
		}

		archesSet := set.From(arches)
		archesFound := set.New[string](0)

		for _, resource := range allResources {
			if resource.Name == fmt.Sprintf("%s/%s", rs.Primary.Attributes["os"], rs.Primary.Attributes["release"]) {
				archesFound.Insert(strings.Split(resource.Architecture, "/")[0])

				resourceDetails, err := conn.BootResource.Get(resource.ID)
				if err != nil {
					return err
				}

				for _, resourceSset := range resourceDetails.Sets {
					if !resourceSset.Complete {
						return fmt.Errorf("resources of the selection are still importing")
					}
				}
			}
		}

		if !archesSet.Equal(archesFound) {
			return fmt.Errorf("selection architectures are missing from imported boot resources")
		}

		*bootSourceSelection = *gotBootSourceSelection

		return nil
	}
}

func testAccMAASBootSourceSelection(os string, release string, arches []string) string {
	archesList, _ := json.Marshal(arches)

	return fmt.Sprintf(`
data "maas_boot_source" "test" {}

resource "maas_boot_source_selection" "test" {
	boot_source = data.maas_boot_source.test.id

	os         = %q
	release    = %q
	arches     = %v
}
`, os, release, string(archesList))
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

		bootSourceID, err := strconv.Atoi(rs.Primary.Attributes["boot_source"])
		if err != nil {
			return err
		}

		response, err := conn.BootSourceSelection.Get(bootSourceID, id)
		if err == nil {
			// default boot source selection leads to noop
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Boot Source Selection (%s %d) still exists.", rs.Primary.ID, bootSourceID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_boot_source_selection is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}

		allResources, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return err
		}

		for _, resource := range allResources {
			if resource.Name == fmt.Sprintf("%s/%s", rs.Primary.Attributes["os"], rs.Primary.Attributes["release"]) {
				return fmt.Errorf("MAAS Boot resources from the deleted selection (%s %d) still exist.", rs.Primary.ID, bootSourceID)
			}
		}
	}

	return nil
}
