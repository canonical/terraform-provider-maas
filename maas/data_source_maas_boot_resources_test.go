package maas_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASBootResources_basic(t *testing.T) {
	os := "ubuntu"
	release := "kinetic"
	arches := []string{"arm64"}
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
		CheckDestroy: testAccCheckDataSourceMAASBootResourcesDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASBootResources(os, release, arches, subarches, labels),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASBootResources(os string, release string, arches []string, subarches []string, labels []string) string {
	return fmt.Sprintf(`
%s

data "maas_boot_resources" "test" {
  os      = maas_boot_source_selection.test.os
  release = maas_boot_source_selection.test.release
}
`, testAccMAASBootSourceSelection(os, release, arches, subarches, labels))
}

func testAccCheckDataSourceMAASBootResourcesDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	err := awaitImportComplete(conn)
	if err != nil {
		return fmt.Errorf("Could not await image importing: %v", err)
	}

	// loop through the resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_boot_resources" {
			continue
		}

		// fetch the boot source
		bootSource, err := conn.BootSources.Get()
		if err != nil {
			return fmt.Errorf("error fetching boot sources: %v", err)
		}

		bootSourceID := bootSource[0].ID

		// fetch all the synced resources
		response, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return fmt.Errorf("error getting synced boot resource: %s", err)
		}

		resourceMap := make(map[string]struct{})
		for _, res := range response {
			resourceMap[res.Name] = struct{}{}
		}

		// we need to read each resource separately
		count := rs.Primary.Attributes["boot_resources.#"]
		selectionCount, err := strconv.Atoi(count)

		if err != nil {
			return fmt.Errorf("Could not convert %v to integer: %v", count, err)
		}

		if selectionCount < 1 {
			return fmt.Errorf("Boot Resource does not contain any selections!")
		}

		// check each resource has been deleted
		for i := range selectionCount {
			thisName := rs.Primary.Attributes[fmt.Sprintf("boot_resources.%d.name", i)]
			if _, exists := resourceMap[thisName]; exists {
				return fmt.Errorf("Boot Resource still exists for %s", thisName)
			}

			parts := strings.SplitN(thisName, "/", 2)
			if len(parts) < 2 {
				return fmt.Errorf("Invalid resource name: %s", thisName)
			}

			os, release := parts[0], parts[1]

			if bootSourceSelection, err := findBootSourceSelection(conn, bootSourceID, os, release); err != nil {
				// 404 means the resource was deleted already
				if !strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				// anything else is an error
				return fmt.Errorf("error finding selection '%v': %v", thisName, err)
			} else if bootSourceSelection != nil {
				return fmt.Errorf("boot source selection (%s) was unexpectedly found on deleted resource", thisName)
			}
		}

		return nil
	}

	return nil
}

func findBootSourceSelection(client *client.Client, bootSource int, os string, release string) (*entity.BootSourceSelection, error) {
	bootSourceSelections, err := client.BootSourceSelections.Get(bootSource)
	if err != nil {
		return nil, err
	}

	for _, d := range bootSourceSelections {
		if d.OS == os && d.Release == release {
			return &d, nil
		}
	}

	return nil, err
}

func awaitImportComplete(client *client.Client) error {
	if err := client.BootResources.Import(); err != nil {
		return err
	}

	timeout := 40 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		if importing, err := client.BootResources.IsImporting(); err != nil {
			return retry.NonRetryableError(err)
		} else if importing {
			return retry.RetryableError(fmt.Errorf("boot resources still importing, waiting... "))
		}

		return nil
	})
	// add a small delay to ensure the resources are fully updated
	if err := retry.RetryContext(ctx, 10*time.Second, func() *retry.RetryError {
		return nil
	}); err != nil {
		return fmt.Errorf("error after waiting 10 seconds: %s", err)
	}

	return result
}
