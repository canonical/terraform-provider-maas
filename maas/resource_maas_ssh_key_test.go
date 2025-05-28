package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASSSHKey_basic(t *testing.T) {
	sshKey, err := generateEd25519Key()
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	checks := []resource.TestCheckFunc{
		testAccCheckMAASSSHKeyExists("maas_ssh_key.test", sshKey),
		resource.TestCheckResourceAttr("maas_ssh_key.test", "key", sshKey),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASSSHKeyDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASSSHKeyConfig(sshKey),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
			{
				ResourceName:      "maas_ssh_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMAASSSHKeyExists(resourceName string, expectedSSHKey string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		
		sshKeyID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error converting SSH key id to int: %v", err)
		}
		sshKeyMAAS, err := client.SSHKey.Get(sshKeyID)
		if err != nil {
			return fmt.Errorf("error getting SSH key with id: %s error: %v", rs.Primary.ID, err)
		}
		if expectedSSHKey != sshKeyMAAS.Key {
			return fmt.Errorf("SSH key does not match expected value")
		}
		return nil
	}
}

func testAccCheckMAASSSHKeyDestroy(s *terraform.State) error {
	client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_ssh_key" {
			continue
		}
		sshKeyId, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error converting SSH key id to int: %v", err)
		}
		response, err := client.SSHKey.Get(sshKeyId)
		if err == nil {
			if response != nil && response.ID == sshKeyId {
				return fmt.Errorf("MAAS SSH Key (%s) still exists.", rs.Primary.ID)
			}
		}
		// If the error is not a 404, the interface has not been destroyed as it should have been
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
		
	}
	return nil
}

func testAccMAASSSHKeyConfig(sshKey string) string {
	return fmt.Sprintf(`
resource "maas_ssh_key" "test" {
  key = %q
}
	`, sshKey)
}

func generateEd25519Key() (string, error) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", err
	}
	sshKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", err
	}
	return string(ssh.MarshalAuthorizedKey(sshKey)), nil
}