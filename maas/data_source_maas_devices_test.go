package maas_test

import (
	"fmt"
	"strconv"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMAASDevices_basic(t *testing.T) {
	checks := []resource.TestCheckFunc{
		func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["data.maas_devices.test"]
			if !ok {
				return fmt.Errorf("data source not found: data.maas_devices.test")
			}

			n, err := strconv.Atoi(rs.Primary.Attributes["devices.#"])
			if err != nil {
				return err
			}

			for i := 0; i < n; i++ {
				if err := resource.TestCheckResourceAttrSet("data.maas_devices.test", fmt.Sprintf("devices.%v.system_id", i))(s); err != nil {
					return err
				}
				if err := resource.TestCheckResourceAttrSet("data.maas_devices.test", fmt.Sprintf("devices.%v.hostname", i))(s); err != nil {
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
				Config: testAccDataSourceMAASDevices(),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASDevices() string {
	return `data "maas_devices" "test" {}`
}
