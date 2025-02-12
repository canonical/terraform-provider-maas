package maas_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceMaasTag_basic(t *testing.T) {

	var tag entity.Tag
	comment := "Test comment"
	name := acctest.RandomWithPrefix("tf-tag-")
	machines := os.Getenv("TF_ACC_TAG_MACHINES")

	checks := []resource.TestCheckFunc{
		testAccMaasTagCheckExists("maas_tag.test", &tag),
		resource.TestCheckResourceAttr("maas_tag.test", "name", name),
		resource.TestCheckResourceAttr("maas_tag.test", "comment", comment),
		resource.TestCheckResourceAttr("maas_tag.test", "machines.#", strconv.Itoa(len(strings.Split(machines, ",")))),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_TAG_MACHINES"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMaasTagDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMaasTag(name, comment, machines),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			// Test import using name
			{
				ResourceName: "maas_tag.test",
				ImportState:  true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					var tag *terraform.InstanceState
					if len(is) != 1 {
						return fmt.Errorf("expected 1 state: %#v", t)
					}
					tag = is[0]
					assert.Equal(t, tag.Attributes["name"], name)
					assert.Equal(t, tag.Attributes["comment"], comment)
					assert.Equal(t, tag.Attributes["machines.#"], strconv.Itoa(len(strings.Split(machines, ","))))
					return nil
				},
			},
		},
	})
}

func testAccMaasTagCheckExists(rn string, tag *entity.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		gotTag, err := conn.Tag.Get(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting tag: %s", err)
		}

		*tag = *gotTag

		return nil
	}
}

func testAccMaasTag(name string, comment string, machines string) string {
	return fmt.Sprintf(`
resource "maas_tag" "test" {
	name        = "%s"
	kernel_opts = "console=tty1 console=ttyS0"
	machines    = split(",", "%s")
	comment     = "%s"
}
`, name, machines, comment)
}

func testAccCheckMaasTagDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_tag
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_tag" {
			continue
		}

		// Retrieve our maas_tag by referencing it's state ID for API lookup
		response, err := conn.Tag.Get(rs.Primary.ID)
		if err == nil {
			if response != nil && response.Name == rs.Primary.ID {
				return fmt.Errorf("MAAS Tag (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_tag is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
