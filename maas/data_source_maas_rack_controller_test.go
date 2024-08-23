package maas_test

import (
	"fmt"
	"regexp"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasRackController_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceMaasRackController("rack-controller"),
				ExpectError: regexp.MustCompile(`rack controller \(rack-controller\) was not found`),
			},
		},
	})
}

func testAccDataSourceMaasRackController(hostname string) string {
	return fmt.Sprintf(`
data "maas_rack_controller" "test" {
	hostname = "%s"
}
`, hostname)
}
