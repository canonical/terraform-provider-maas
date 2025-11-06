package maas_test

import (
	"testing"

	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMAASRackControllers_basic(t *testing.T) {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("data.maas_rack_controllers.test", "controllers.#"),
		resource.TestCheckResourceAttrSet("data.maas_rack_controllers.test", "controllers.0.hostname"),
		resource.TestCheckResourceAttrSet("data.maas_rack_controllers.test", "controllers.0.id"),
		resource.TestCheckResourceAttrSet("data.maas_rack_controllers.test", "controllers.0.services.#"),
		resource.TestCheckResourceAttrSet("data.maas_rack_controllers.test", "controllers.0.subnets.#"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASRackControllers(),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASRackControllers() string {
	return `
data "maas_rack_controllers" "test" {}
`
}
