package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasResourcePool_basic(t *testing.T) {

	var resourcePool entity.ResourcePool
	description := "Test description"
	name := acctest.RandomWithPrefix("tf-resource-pool-")

	checks := []resource.TestCheckFunc{
		testAccMaasResourcePoolCheckExists("maas_resource_pool.test", &resourcePool),
		resource.TestCheckResourceAttr("data.maas_resource_pool.test", "description", description),
		resource.TestCheckResourceAttr("data.maas_resource_pool.test", "name", name),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasResourcePoolDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasResourcePool(description, name),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasResourcePool(description string, name string) string {
	return fmt.Sprintf(`
%s

data "maas_resource_pool" "test" {
	name = maas_resource_pool.test.name
}
`, testAccMaasResourcePool(description, name))
}
