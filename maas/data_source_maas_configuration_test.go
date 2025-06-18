package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMAASConfiguration_basic(t *testing.T) {
	key := "remote_syslog"
	value := "example.com:514"
	value2 := "" // Set back to default.

	composeCheckFuncs := func(val string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			testAccMAASConfigurationCheckExists("maas_configuration.test"),
			resource.TestCheckResourceAttr("data.maas_configuration.test", "key", key),
			resource.TestCheckResourceAttr("data.maas_configuration.test", "value", val),
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccMAASConfigurationCheckDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASConfiguration(key, value),
				Check:  composeCheckFuncs(value),
			},
			{
				Config: testAccDataSourceMAASConfiguration(key, value2),
				Check:  composeCheckFuncs(value2),
			},
		},
	})
}

func testAccDataSourceMAASConfiguration(key string, value string) string {
	return fmt.Sprintf(`
%s

data "maas_configuration" "test" {
	key = maas_configuration.test.key
}
`, testAccMAASConfigurationConfigBasic(key, value))
}
