package maas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceMaasTag(t *testing.T) {

	name := acctest.RandomWithPrefix("tf-test-")

	tf := fmt.Sprintf(testAccResourceMaasTagConfig, name)
	resourceName := "maas_tag.tf_tag"

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

const testAccResourceMaasTagConfig = `
resource "maas_tag" "tf_tag" {
  name = "%s"
}
`
