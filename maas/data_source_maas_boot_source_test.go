package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootSource_basic(t *testing.T) {

	url := "https://images.maas.io/ephemeral-v3/candidate/"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.maas_boot_source.test", "url", url),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "created"),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "updated"),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "keyring_data"),
		resource.TestCheckResourceAttrSet("data.maas_boot_source.test", "keyring_filename"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootSource(url),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootSource(url string) string {
	return fmt.Sprintf(`

data "maas_boot_source" "test" {
	url = "%s"
}
`, url)
}
