package maas_test

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"terraform-provider-maas/maas"
	"testing"

	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBlockDeviceTag_basic(t *testing.T) {
	machine := os.Getenv("TF_ACC_BLOCK_DEVICE_MACHINE")

	blockDeviceName := acctest.RandomWithPrefix("tf")
	tagName := acctest.RandomWithPrefix("tag")
	tagName2 := acctest.RandomWithPrefix("tag")
	tagName3 := acctest.RandomWithPrefix("tag")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccCheckMaasBlockDeviceTagDestroy,
		Steps: []resource.TestStep{
			// Test create.
			{
				Config: testAccBlockDeviceTagConfig(machine, blockDeviceName, tagName, tagName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaasBlockDeviceTagExists("maas_block_device_tag.test", tagName, tagName2),
					resource.TestCheckResourceAttr("maas_block_device_tag.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("maas_block_device_tag.test", "tags.*", tagName),
					resource.TestCheckTypeSetElemAttr("maas_block_device_tag.test", "tags.*", tagName2),
				),
			},
			// Test update. Expected behaviour is that the previous tag is removed and the new tag is added.
			{
				Config: testAccBlockDeviceTagConfig(machine, blockDeviceName, tagName2, tagName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaasBlockDeviceTagExists("maas_block_device_tag.test", tagName2, tagName3),
					resource.TestCheckResourceAttr("maas_block_device_tag.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("maas_block_device_tag.test", "tags.*", tagName2),
					resource.TestCheckTypeSetElemAttr("maas_block_device_tag.test", "tags.*", tagName3),
				),
			},
			// Test import
			{
				ResourceName:      "maas_block_device_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_block_device_tag.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: maas_block_device_tag.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["machine"], rs.Primary.Attributes["block_device_id"]), nil
				},
			},
		},
	})
}

func testAccBlockDeviceTagConfig(hostname string, name string, tagNames ...string) string {
	return fmt.Sprintf(`

data "maas_machine" "machine" {
  hostname = %q
}

resource "maas_block_device" "test" {
  machine        = data.maas_machine.machine.id
  name           = %q
  size_gigabytes = 1
  id_path        = "/dev/test"
}

resource "maas_block_device_tag" "test" {
  block_device_id = maas_block_device.test.id
  machine         = maas_block_device.test.machine
  tags            = %s
}
	`, hostname, name, fmt.Sprintf("[\"%s\"]", strings.Join(tagNames, "\", \"")))
}

func testAccCheckMaasBlockDeviceTagExists(resourceName string, tagNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		systemId, blockDeviceId, err := maas.SplitTagStateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		blockDevice, err := conn.BlockDevice.Get(systemId, blockDeviceId)
		if err != nil {
			return err
		}

		// Check the block device is the one expected
		if blockDevice.ID != blockDeviceId {
			return fmt.Errorf("MAAS Block Device (%v) ID mismatch: expected %v, got %v.", blockDevice.ID, blockDeviceId, blockDevice.ID)
		}

		// Check the tags exist
		for _, tag := range tagNames {
			if !slices.Contains(blockDevice.Tags, tag) {
				return fmt.Errorf("MAAS Block Device (%d) tags (%s) do not match expected tags: %s", blockDevice.ID, blockDevice.Tags, tagNames)
			}
		}

		return nil
	}
}

func testAccCheckMaasBlockDeviceTagDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_block_device_tag is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_block_device_tag" {
			continue
		}

		// Retrieve the system and block device ID from the state ID
		systemId, blockDeviceId, err := maas.SplitTagStateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Check the block device doesn't exist
		response, err := conn.BlockDevice.Get(systemId, blockDeviceId)
		if err == nil {
			if response != nil && response.ID == blockDeviceId {
				return fmt.Errorf("MAAS Block Device (%s) still exists.", rs.Primary.ID)
			}
		}
		// If the error is a 404, the block device is destroyed as expected
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}
	return nil
}
