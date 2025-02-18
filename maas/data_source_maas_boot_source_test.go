package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootSource_basic(t *testing.T) {

	url := "http://images.maas.io/ephemeral-v3/stable/"
	keyring_path := "/snap/maas/current/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.maas_boot_source.test", "url", url),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "created"),
		resource.TestCheckResourceAttr("data.maas_boot_source.test", "keyring_data", ""),
		resource.TestCheckResourceAttr("data.maas_boot_source.test", "keyring_filename", keyring_path),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "updated"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootSource(url, keyring_path),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootSource(url string, keyring_path string) string {
	return fmt.Sprintf(`
%s

data "maas_boot_source" "test" {}`, testAccMAASBootSource(url, keyring_path))
}
