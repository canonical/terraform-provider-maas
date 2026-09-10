package maas_test

import (
	"fmt"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASAPIKey_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-api-key-")
	nameUpdated := acctest.RandomWithPrefix("tf-api-key-updated-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckAPIKeyDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: testAccMAASAPIKey(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists("maas_api_key.test"),
					resource.TestCheckResourceAttr("maas_api_key.test", "name", name),
					resource.TestCheckResourceAttrSet("maas_api_key.test", "token_key"),
					resource.TestCheckResourceAttrSet("maas_api_key.test", "token_secret"),
					resource.TestCheckResourceAttrSet("maas_api_key.test", "consumer_key"),
					resource.TestCheckResourceAttrSet("maas_api_key.test", "api_key"),
				),
			},
			// Test update (rename)
			{
				Config: testAccMAASAPIKey(nameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists("maas_api_key.test"),
					resource.TestCheckResourceAttr("maas_api_key.test", "name", nameUpdated),
					resource.TestCheckResourceAttrSet("maas_api_key.test", "api_key"),
				),
			},
			// Test import
			{
				ResourceName:      "maas_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMAASAPIKey(name string) string {
	return fmt.Sprintf(`
resource "maas_api_key" "test" {
	name = %q
}
`, name)
}

func testAccCheckAPIKeyExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		tokens, err := conn.Account.ListAuthorisationTokens()
		if err != nil {
			return fmt.Errorf("error listing authorisation tokens: %s", err)
		}

		for _, t := range tokens {
			parts := strings.SplitN(t.Token, ":", 3)
			if len(parts) >= 2 && parts[1] == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("API key with token_key %q not found", rs.Primary.ID)
	}
}

func testAccCheckAPIKeyDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_api_key" {
			continue
		}

		tokenKey := rs.Primary.ID

		tokens, err := conn.Account.ListAuthorisationTokens()
		if err != nil {
			return fmt.Errorf("error listing authorisation tokens: %s", err)
		}

		for _, t := range tokens {
			parts := strings.SplitN(t.Token, ":", 3)
			if len(parts) >= 2 && parts[1] == tokenKey {
				return fmt.Errorf("API key with token_key %q still exists", tokenKey)
			}
		}
	}

	return nil
}
