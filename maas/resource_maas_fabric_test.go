package maas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceMaasFabric(t *testing.T) {

	name := acctest.RandomWithPrefix("tf-test-") // this is to avoid name conflicts if multiple tests are running

	tf := fmt.Sprintf(testAccResourceMaasFabricConfig, name)
	resourceName := "maas_fabric.tf_fabric"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          nil,
		ProviderFactories: providerFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
		},
	})
}

const testAccResourceMaasFabricConfig = `
resource "maas_fabric" "tf_fabric" {
	name = "%s"
  }
`
