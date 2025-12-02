package maas_test

import (
	"fmt"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASUser_basic(t *testing.T) {
	username1 := "testUser1"
	password1 := "password1"
	email1 := "testuser1@email.com"

	username2 := "testUser2"
	password2 := "password2"
	email2 := "testuser2@email.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckUserDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test initial creation
			{
				Config: testAccUser(username1, password1, email1, true) + testAccUserTransfer(username2, password2, email2, false, username1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("maas_user.test_testUser1"),
					testAccCheckUserExists("maas_user.test_testUser2"),

					resource.TestCheckResourceAttr("maas_user.test_testUser1", "name", username1),
					resource.TestCheckResourceAttr("maas_user.test_testUser1", "is_admin", fmt.Sprintf("%t", true)),
					resource.TestCheckResourceAttr("maas_user.test_testUser1", "email", email1),
					resource.TestCheckResourceAttr("maas_user.test_testUser1", "password", password1),
					resource.TestCheckNoResourceAttr("maas_user.test_testUser1", "transfer_to_user"),

					resource.TestCheckResourceAttr("maas_user.test_testUser2", "name", username2),
					resource.TestCheckResourceAttr("maas_user.test_testUser2", "is_admin", fmt.Sprintf("%t", false)),
					resource.TestCheckResourceAttr("maas_user.test_testUser2", "email", email2),
					resource.TestCheckResourceAttr("maas_user.test_testUser2", "password", password2),
					resource.TestCheckResourceAttr("maas_user.test_testUser2", "transfer_to_user", username1),
				),
			},
			// Test deletion with transfer
			{
				Config: testAccUser(username1, password1, email1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("maas_user.test_testUser1"),

					resource.TestCheckResourceAttr("maas_user.test_testUser1", "name", username1),
					resource.TestCheckResourceAttr("maas_user.test_testUser1", "is_admin", fmt.Sprintf("%t", true)),
					resource.TestCheckResourceAttr("maas_user.test_testUser1", "email", email1),
					resource.TestCheckResourceAttr("maas_user.test_testUser1", "password", password1),
					resource.TestCheckNoResourceAttr("maas_user.test_testUser1", "transfer_to_user"),
				),
			},
			// Test import
			{
				ResourceName:            "maas_user.test_testUser1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_user.test_testUser1"]
					if !ok {
						return "", fmt.Errorf("resource not found: maas_user.test_testUser1")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}

					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccUser(username string, password string, email string, isAdmin bool) string {
	return fmt.Sprintf(`
resource "maas_user" "test_%v" {
  name     = %q
  password = %q
  email    = %q
  is_admin = %t
}`, username, username, password, email, isAdmin)
}

func testAccUserTransfer(username string, password string, email string, isAdmin bool, transfer string) string {
	return fmt.Sprintf(`
resource "maas_user" "test_%v" {
  name     = %q
  password = %q
  email    = %q
  is_admin = %t

  transfer_to_user = %q
}`, username, username, password, email, isAdmin, transfer)
}

func testAccCheckUserExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		_, err := conn.User.Get(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting the user: %s", err)
		}

		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_user" {
			continue
		}

		id := rs.Primary.ID

		response, err := conn.User.Get(id)
		if err == nil {
			if response != nil && response.UserName == id {
				return fmt.Errorf("User %q still exists.", response.UserName)
			}
		}

		// 404 means destroyed, anything else is an error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
