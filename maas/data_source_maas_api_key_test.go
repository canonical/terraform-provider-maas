package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMAASAPIKey_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-api-key-ds-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASAPIKey(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.maas_api_key.test", "name", name),
					resource.TestCheckResourceAttrSet("data.maas_api_key.test", "token_key"),
					resource.TestCheckResourceAttrSet("data.maas_api_key.test", "token_secret"),
					resource.TestCheckResourceAttrSet("data.maas_api_key.test", "consumer_key"),
					resource.TestCheckResourceAttrSet("data.maas_api_key.test", "api_key"),
					resource.TestCheckResourceAttrPair(
						"data.maas_api_key.test", "api_key",
						"maas_api_key.test", "api_key",
					),
				),
			},
		},
	})
}

func testAccDataSourceMAASAPIKey(name string) string {
	return fmt.Sprintf(`
resource "maas_api_key" "test" {
	name = %q
}

data "maas_api_key" "test" {
	name = maas_api_key.test.name
}
`, name)
}
