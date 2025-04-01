package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMAASZone_basic(t *testing.T) {
	var zone entity.Zone

	description := "Test description"
	name := acctest.RandomWithPrefix("tf-zone-")

	checks := []resource.TestCheckFunc{
		testAccMAASZoneCheckExists("maas_zone.test", &zone),
		resource.TestCheckResourceAttr("data.maas_zone.test", "description", description),
		resource.TestCheckResourceAttr("data.maas_zone.test", "name", name),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASZoneDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASZone(description, name),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMAASZone(description string, name string) string {
	return fmt.Sprintf(`
%s

data "maas_zone" "test" {
	name = maas_zone.test.name
}
`, testAccMAASZone(name, description))
}
