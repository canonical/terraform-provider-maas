package maas_test

import (
	"fmt"
	"strconv"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASMachines_basic(t *testing.T) {
	checks := []resource.TestCheckFunc{
		func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["data.maas_machines.test"]
			if !ok {
				return fmt.Errorf("data source not found: data.maas_machines.test")
			}

			n, err := strconv.Atoi(rs.Primary.Attributes["machines.#"])
			if err != nil {
				return err
			}

			for i := 0; i < n; i++ {
				if err := resource.TestCheckResourceAttrSet("data.maas_machines.test", fmt.Sprintf("machines.%v.system_id", i))(s); err != nil {
					return err
				}
				if err := resource.TestCheckResourceAttrSet("data.maas_machines.test", fmt.Sprintf("machines.%v.hostname", i))(s); err != nil {
					return err
				}
			}

			return nil
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASMachines(),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASMachines() string {
	return `data "maas_machines" "test" {}`
}
