package maas_test

import (
	"fmt"

	"reflect"
	"slices"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"crypto/ed25519"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/crypto/ssh"
)

func TestSplitSSHKeyStateID(t *testing.T) {
	testCases := []struct {
		id       string
		expected []int
	}{
		{
			id:       "1/2/3",
			expected: []int{1, 2, 3},
		},
		{
			id:       "1",
			expected: []int{1},
		},
	}

	for _, testCase := range testCases {
		actual, err := maas.SplitSSHKeyStateID(testCase.id)
		if err != nil {
			t.Fatalf("error splitting SSH key state ID: %v", err)
		}

		if !reflect.DeepEqual(actual, testCase.expected) {
			t.Fatalf("expected %v, got %v", testCase.expected, actual)
		}
	}
}

func TestCreateIDFromKeys(t *testing.T) {
	testCases := []struct {
		keys     []entity.SSHKey
		expected string
	}{
		{
			keys:     []entity.SSHKey{{ID: 1}, {ID: 2}, {ID: 3}},
			expected: "1/2/3",
		},
		{
			keys:     []entity.SSHKey{{ID: 10}},
			expected: "10",
		},
	}
	for _, testCase := range testCases {
		actual := maas.CreateIDFromKeys(testCase.keys)
		if actual != testCase.expected {
			t.Fatalf("expected %v, got %v", testCase.expected, actual)
		}
	}
}

func TestAccResourceMAASSSHKey_basic(t *testing.T) {
	sshKey1, err := generateEd25519Key()
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	sshKey2, err := generateEd25519Key()
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	singleKey := []string{sshKey1}
	sshKeys := []string{sshKey1, sshKey2}

	singleKeyChecks := []resource.TestCheckFunc{
		testAccCheckMAASSSHKeyExists("maas_ssh_keys.test", singleKey),
		resource.TestCheckResourceAttr("maas_ssh_keys.test", "keys.#", "1"),
		resource.TestCheckTypeSetElemAttr("maas_ssh_keys.test", "keys.*", sshKey1),
	}
	multiKeyChecks := []resource.TestCheckFunc{
		testAccCheckMAASSSHKeyExists("maas_ssh_keys.test", sshKeys),
		resource.TestCheckResourceAttr("maas_ssh_keys.test", "keys.#", "2"),
		resource.TestCheckTypeSetElemAttr("maas_ssh_keys.test", "keys.*", sshKey1),
		resource.TestCheckTypeSetElemAttr("maas_ssh_keys.test", "keys.*", sshKey2),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASSSHKeyDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASSSHKeyConfig(singleKey),
				Check:  resource.ComposeTestCheckFunc(singleKeyChecks...),
			},
			{
				ResourceName:      "maas_ssh_keys.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMAASSSHKeyConfig(sshKeys),
				Check:  resource.ComposeTestCheckFunc(multiKeyChecks...),
			},
			{
				ResourceName:      "maas_ssh_keys.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMAASSSHKeyExists(resourceName string, expectedSSHKeys []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		allKeys, err := maas.SplitSSHKeyStateID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting SSH key state ID: %v", err)
		}

		for _, sshKeyID := range allKeys {
			sshKeyMAAS, err := client.SSHKey.Get(sshKeyID)
			if err != nil {
				return fmt.Errorf("error getting SSH key with id: %s error: %v", rs.Primary.ID, err)
			}

			if !slices.Contains(expectedSSHKeys, sshKeyMAAS.Key) {
				return fmt.Errorf("SSH key does not match expected value")
			}
		}

		return nil
	}
}

func testAccCheckMAASSSHKeyDestroy(s *terraform.State) error {
	client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_ssh_keys" {
			continue
		}

		sshKeyIDs, err := maas.SplitSSHKeyStateID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting SSH key state ID: %v", err)
		}

		for _, sshKeyID := range sshKeyIDs {
			response, err := client.SSHKey.Get(sshKeyID)
			if err == nil {
				if response != nil && response.ID == sshKeyID {
					return fmt.Errorf("MAAS SSH Key (%d) still exists.", sshKeyID)
				}
			}
			// If the error is not a 404, the interface has not been destroyed as it should have been
			if !strings.Contains(err.Error(), "404 Not Found") {
				return err
			}
		}
	}

	return nil
}

func testAccMAASSSHKeyConfig(sshKeys []string) string {
	return fmt.Sprintf(`
resource "maas_ssh_keys" "test" {
  keys = %v
}
	`, testutils.StringifySliceAsLiteralArray(sshKeys))
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

	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshKey))), nil
}
