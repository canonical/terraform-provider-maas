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
	"github.com/canonical/gomaasclient/entity"
	"github.com/canonical/gomaasclient/entity/node"

)

func TestAccResourceMAASInstance_basic(t *testing.T) {
	// Required to check if the machine is released into its read state after destroy
	var machine entity.Machine  

	vm_host := os.Getenv("TF_ACC_VM_HOST_ID")
	hostname := acctest.RandomWithPrefix("tf-instance")
	comment := acctest.RandomWithPrefix("tf-instance-comment")
	erase := "true"
	force := "false"
	quick_erase := "true"
	secure_erase := "false"


	baseChecks := []resource.TestCheckFunc{
		testAccMAASInstanceCheckExists("maas_instance.test", &machine),
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
				Config: testAccMAASInstanceConfigBasic(vm_host, hostname),
				Check: resource.ComposeTestCheckFunc(baseChecks...),
			},
			// Test update
			{
				Config: testAccMAASInstanceConfigSetup(vm_host, hostname)  + testAccMAASInstanceConfigReleaseParams(comment, erase, force, quick_erase, secure_erase),
				Check: resource.ComposeTestCheckFunc(append(
					baseChecks, 
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.#", "1"),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.comment", comment),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.erase", erase),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.force", force),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.quick_erase", quick_erase),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.secure_erase", secure_erase),
				)...,
				),
			},
			// Test destroy leaves the machine in a ready state
			{
				Config: testAccMAASInstanceConfigSetup(vm_host, hostname),
				Check: testAccMAASInstanceCheckMachineInStatus(machine.SystemID, node.StatusReady),
			},
		},
	})
}

func testAccMAASInstanceCheckMachineInStatus(systemID string, status node.Status) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		machine, err := conn.Machine.Get(systemID)
		if err != nil {
			return err
		}

		if machine.Status != status {
			return fmt.Errorf("machine %s is not in the expected status %d but in status %d", systemID, status, machine.Status)
		}

		return nil
	}
}


func testAccMAASInstanceCheckExists(rn string, machine *entity.Machine) resource.TestCheckFunc {
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
		gotMachine, err := conn.Machine.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		if gotMachine.SystemID != rs.Primary.ID {
			return fmt.Errorf("machine ID %s does not match expected id %s", gotMachine.SystemID, rs.Primary.ID)
		}

		*machine = *gotMachine

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

func testAccMAASInstanceConfigSetup(vm_host, hostname string) string {
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

func testAccMAASInstanceConfigBasic(vm_host, hostname string) string {
	return fmt.Sprintf(`
%s 

resource "maas_instance" "test" {
  allocate_params {
    hostname      = maas_vm_host_machine.test.hostname
    min_memory    = 4000
    min_cpu_count = 1
  }
}
`, testAccMAASInstanceConfigSetup(vm_host, hostname),)
}


func testAccMAASInstanceConfigReleaseParams(comment, erase, force, quick_erase, secure_erase string) string {
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
    min_memory    = 4000
    min_cpu_count = 1
  }

}
`, comment, erase, force, quick_erase, secure_erase)
	}
