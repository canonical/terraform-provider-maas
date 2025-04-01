package maas_test

import (
	"fmt"
	"slices"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASBootSource_basic(t *testing.T) {
	keyringPath := "/snap/maas/current/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"
	imageURLs := []string{
		"http://images.maas.io/ephemeral-v3/stable/",
		"http://images.maas.io/ephemeral-v3/candidate/",
	}

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrWith("data.maas_boot_source.test", "url", func(value string) error {
			if !slices.Contains(imageURLs, value) {
				return fmt.Errorf("expected to be one of %v, got %v", imageURLs, value)
			}

			return nil
		}),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "created"),
		resource.TestCheckResourceAttr("data.maas_boot_source.test", "keyring_data", ""),
		resource.TestCheckResourceAttr("data.maas_boot_source.test", "keyring_filename", keyringPath),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "updated"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASBootSource(),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASBootSource() string {
	return `data "maas_boot_source" "test" {}`
}
