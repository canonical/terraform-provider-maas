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
		CheckDestroy: testAccCheckDataSourceMaasBootResourcesDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootResources(os, release, arches, subarches, labels),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootResources(os string, release string, arches []string, subarches []string, labels []string) string {
	return fmt.Sprintf(`
%s

data "maas_boot_resources" "test" {
	os      = maas_boot_source_selection.test.os
	release = maas_boot_source_selection.test.release
}
`, testAccMAASBootSourceSelection(os, release, arches, subarches, labels))
}

func testAccCheckDataSourceMaasBootResourcesDestroy(s *terraform.State) error {
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

		response, err := conn.BootResources.Get(&entity.BootResourcesReadParams{Type: "synced"})
		if err != nil {
			return fmt.Errorf("error getting synced boot resource: %s", err)
		}
		resourceMap := make(map[string]struct{})
		for _, res := range response {
			resourceMap[res.Name] = struct{}{}
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		bootsource, err := conn.BootSources.Get()
		if err != nil {
			return fmt.Errorf("error fetching boot sources: %v", err)
		}
		boot_source_id := bootsource[0].ID

		// we need to read each resource seperately
		count := rs.Primary.Attributes["boot_resources.#"]
		selectionCount, err := strconv.Atoi(count)
		if err != nil {
			return fmt.Errorf("Could not convert %v to integer: %v", count, err)
		}
		if selectionCount < 1 {
			return fmt.Errorf("Boot Resource does not contain any selections!")
		}

		for i := 0; i < selectionCount; i++ {
			this_name := rs.Primary.Attributes[fmt.Sprintf("boot_resources.%d.name", i)]
			if _, exists := resourceMap[this_name]; exists {
				return fmt.Errorf("Boot Resource still exists for %s", this_name)
			}

			parts := strings.SplitN(this_name, "/", 2)
			if len(parts) < 2 {
				return fmt.Errorf("Invalid resource name: %s", this_name)
			}
			os, release := parts[0], parts[1]

			if bootsourceselection, err := findBootSourceSelection(conn, boot_source_id, os, release); err != nil {
				// 404 means the resource was deleted already
				if !strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				// anything else is an error
				return fmt.Errorf("error finding selection '%v': %v", this_name, err)
			} else if bootsourceselection != nil {
				return fmt.Errorf("boot source selection (%s) was unexpectedly found on deleted resource", this_name)
			}
		}
		return nil
	}

	return nil
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
	// wait for everything to take effect
	time.Sleep(10 * time.Second)
	return result
}

func findBootSourceSelection(client *client.Client, boot_source int, os string, release string) (*entity.BootSourceSelection, error) {
	if bootsourceselections, err := client.BootSourceSelections.Get(boot_source); err != nil {
		return nil, err
	} else {
		for _, d := range bootsourceselections {
			if d.OS == os && d.Release == release {
				return &d, nil
			}
		}
	}
	return nil, nil
}
