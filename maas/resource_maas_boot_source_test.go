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

const defaultURL = "http://images.maas.io/ephemeral-v3/stable/"

// We assume tests are running from a snap MAAS environment
const snapKeyring = "/snap/maas/current/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"

func TestAccResourceMAASBootSource_basic(t *testing.T) {

	var bootsource entity.BootSource
	url := "http://images.maas.io/ephemeral-v3/candidate/"

	checks := []resource.TestCheckFunc{
		testAccMAASBootSourceCheckExists("maas_boot_source.test", &bootsource),
		resource.TestCheckResourceAttr("maas_boot_source.test", "url", url),
		resource.TestCheckResourceAttr("maas_boot_source.test", "keyring_filename", snapKeyring),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASBootSource(url, snapKeyring),
				Check:  resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASBootSourceCheckExists(rn string, bootSource *entity.BootSource) resource.TestCheckFunc {
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
		gotBootSource, err := conn.BootSource.Get(id)
		if err != nil {
			return fmt.Errorf("error getting boot source: %s", err)
		}

		*bootSource = *gotBootSource

		return nil
	}
}

func testAccMAASBootSource(url string, keyring_filename string) string {
	return fmt.Sprintf(`
resource "maas_boot_source" "test" {
	url              = "%s"
	keyring_filename = "%s"
}`, url, keyring_filename)
}

func testAccCheckMAASBootSourceDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_boot_source" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		response, err := conn.BootSource.Get(id)
		if err == nil {
			if response.URL != defaultURL {
				return fmt.Errorf("MAAS Boot Source (%s) not reset to default. Returned value: %s", rs.Primary.ID, response.URL)
			}
			if response.KeyringFilename != snapKeyring {
				return fmt.Errorf("MAAS Boot Source (%s) not reset to default. Returned value: %s", rs.Primary.ID, response.KeyringFilename)
			}
			if response.KeyringData != "" {
				return fmt.Errorf("MAAS Boot Source (%s) not reset to default. Returned value: %s", rs.Primary.ID, response.KeyringData)
			}

			return nil
		}

		return err
	}

	return nil
}
