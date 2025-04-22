package maas_test

import (
	// "encoding/json"
	"fmt"
	"os"

	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASInstance_basic(t *testing.T) {
	vm_host := os.Getenv("TF_ACC_VM_HOST_ID")
	hostname := acctest.RandomWithPrefix("tf-instance")
	comment := acctest.RandomWithPrefix("tf-instance-comment")
	erase := "true"
	force := "false"
	quick_erase := "true"
	secure_erase := "false"

	checks := []resource.TestCheckFunc{
		testAccMAASInstanceCheckExists("maas_instance.test"),
		resource.TestCheckResourceAttr("maas_instance.test", "release_params.#", "1"),
		resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.comment", comment),
		resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.erase", erase),
		resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.force", force),
		resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.quick_erase", quick_erase),
		resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.secure_erase", secure_erase),
		resource.TestCheckResourceAttr("maas_instance.test", "hostname", hostname),
		resource.TestCheckResourceAttr("maas_instance.test", "memory", "4096"),
		resource.TestCheckResourceAttr("maas_instance.test", "cpu_count", "1"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"})},
		Providers: testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccMAASInstanceConfigSetup(vm_host, hostname) + testAccMAASInstanceConfig(comment, erase, force, quick_erase, secure_erase),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASInstanceCheckExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the terraform resource from state
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("not found: %s", rn)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set: %s", rn)
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		// Check if the resource exists in MAAS
		machine, err := conn.Machine.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		if machine.SystemID != rs.Primary.ID {
			return fmt.Errorf("machine ID %s does not match expected id %s", machine.SystemID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccMAASInstanceCheckDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_instance" {
			continue
		}

		// Get the machine from maas and check if it exists
		response, err := conn.Machine.Get(rs.Primary.ID)
		if err == nil {
			if response != nil && response.SystemID == rs.Primary.ID {
				return fmt.Errorf("instance %s still exists", rs.Primary.ID)
			}
		} 

		// If the error is equivalent to a 404, the instance was destroyed as expected.
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}

func testAccMAASInstanceConfigSetup(vm_host string, hostname string) string {
	return fmt.Sprintf(
`
resource "maas_vm_host_machine" "test" {
  vm_host  = %q
  cores    = 1
  memory   = 4096  # set to above the default
  hostname = %q
}

`, vm_host, hostname)
	}

func testAccMAASInstanceConfig(comment, erase, force, quick_erase, secure_erase string) string {
	return fmt.Sprintf(
`
resource "maas_instance" "test" {
  release_params {
    comment      = %q
    erase        = %q
    force        = %q
    quick_erase  = %q
    secure_erase = %q	
  }

  allocate_params {
    hostname      = maas_vm_host_machine.test.hostname
    min_memory    = 4096
    min_cpu_count = 1
  }

}
`, comment, erase, force, quick_erase, secure_erase)
	}
