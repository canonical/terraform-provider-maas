package maas_test

import (
	"fmt"
	"os"
	"regexp"

	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASInstance_releaseScripts(t *testing.T) {
	testutils.SkipTestIfNotMAASVersion(t, ">=3.5.0")

	vmHost := os.Getenv("TF_ACC_VM_HOST_ID")
	hostname := acctest.RandomWithPrefix("tf-instance")
	scriptName := acctest.RandomWithPrefix("tf-release-script")

	baseChecks := []resource.TestCheckFunc{
		testAccMAASInstanceCheckExists("maas_instance.test"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccMAASInstanceCheckDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccMAASInstanceConfigBasic(vmHost, hostname) + testAccMAASInstanceConfigReleaseScriptSetup(scriptName),
				Check:  resource.ComposeTestCheckFunc(baseChecks...),
			},
			// Test update
			{
				Config: testAccMAASInstanceConfigSetup(vmHost, hostname) + testAccMAASInstanceConfigReleaseScriptSetup(scriptName) + testAccMAASInstanceConfigReleaseScripts(),
				Check: resource.ComposeTestCheckFunc(append(
					baseChecks,
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.#", "1"),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.scripts.#", "1"),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.scripts.0", scriptName),
				)...,
				),
			},
			// Test destroy releases the machine and runs the release script
			{
				Config: testAccMAASInstanceConfigSetup(vmHost, hostname) + testAccMAASInstanceConfigReleaseScriptSetup(scriptName),
				Check:  testAccMAASInstanceCheckReleaseScriptsRanOnDestroy(hostname, scriptName),
			},
		},
	})
}

func TestAccResourceMAASInstance_basic(t *testing.T) {
	vmHost := os.Getenv("TF_ACC_VM_HOST_ID")
	hostname := acctest.RandomWithPrefix("tf-instance")
	comment := acctest.RandomWithPrefix("tf-instance-comment")
	erase := "true"
	force := "false"
	quickErase := "true"
	secureErase := "false"

	nonDefaultArchitecture := "arm64/generic"
	archError := regexp.MustCompile("Architecture not recognised")

	baseChecks := []resource.TestCheckFunc{
		testAccMAASInstanceCheckExists("maas_instance.test"),
		resource.TestCheckResourceAttr("maas_instance.test", "hostname", hostname),
		resource.TestCheckResourceAttr("maas_instance.test", "memory", "4096"),
		resource.TestCheckResourceAttr("maas_instance.test", "cpu_count", "1"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:    testutils.TestAccProviders,
		ErrorCheck:   func(err error) error { return err },
		CheckDestroy: testAccMAASInstanceCheckDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccMAASInstanceConfigBasic(vmHost, hostname),
				Check: resource.ComposeTestCheckFunc(append(
					baseChecks,
					resource.TestCheckResourceAttr("maas_instance.test", "architecture", "amd64/generic"),
				)...,
				),
			},
			// Test different architecture
			{
				Config:      testAccMAASInstanceConfigAllocateParamsNonDefaultArchitecture(vmHost, hostname, nonDefaultArchitecture),
				ExpectError: archError,
			},
			// Test update
			{
				Config: testAccMAASInstanceConfigSetup(vmHost, hostname) + testAccMAASInstanceConfigReleaseParams(comment, erase, force, quickErase, secureErase),
				Check: resource.ComposeTestCheckFunc(append(
					baseChecks,
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.#", "1"),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.comment", comment),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.erase", erase),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.force", force),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.quick_erase", quickErase),
					resource.TestCheckResourceAttr("maas_instance.test", "release_params.0.secure_erase", secureErase),
				)...,
				),
			},
			// Test destroy leaves the machine in a ready state
			{
				Config: testAccMAASInstanceConfigSetup(vmHost, hostname),
				Check:  testAccMAASInstanceCheckMachineLogsForDestroy(hostname, erase == "true"),
			},
		},
	})
}

// Check logs for relevant events to determine if the machine was released as expected during destroy
func testAccMAASInstanceCheckMachineLogsForDestroy(hostname string, erase bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		params := entity.EventParams{
			Hostname: hostname,
		}

		events, err := conn.Events.Get(&params)
		if err != nil {
			return err
		}

		if len(events.Events) == 0 {
			return fmt.Errorf("no events found for hostname %s", hostname)
		}

		// Check through all events to see if the machine was released as expected
		wasErased := false
		wasReleased := false

		for _, event := range events.Events {
			if event.Type == "Disks erased" {
				wasErased = true
			}

			if event.Type == "Released" {
				wasReleased = true
			}
		}

		if !wasReleased {
			return fmt.Errorf("machine %s was not released as expected", hostname)
		}

		if wasErased != erase {
			return fmt.Errorf("machine %s did not have disks erased as expected", hostname)
		}

		return nil
	}
}

func testAccMAASInstanceCheckReleaseScriptsRanOnDestroy(hostname, scriptName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		releaseScriptRan := false

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "maas_vm_host_machine" {
				continue
			}

			if rs.Primary.Attributes["hostname"] != hostname {
				continue
			}

			params := entity.NodeResultParams{
				Type: entity.RELEASE,
			}

			results, err := conn.NodeResults.Get(rs.Primary.ID, &params)
			if err != nil {
				return err
			}

			for _, result := range results {
				for _, item := range result.Results {
					if item.Name == scriptName {
						releaseScriptRan = true
					}
				}
			}
		}

		if !releaseScriptRan {
			return fmt.Errorf("release script did not run for machine %s", scriptName)
		}

		return nil
	}
}

func testAccMAASInstanceCheckExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set: %s", rn)
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		gotMachine, err := conn.Machine.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if gotMachine.SystemID != rs.Primary.ID {
			return fmt.Errorf("machine ID %s does not match expected id %s", gotMachine.SystemID, rs.Primary.ID)
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

func testAccMAASInstanceConfigSetup(vmHost, hostname string) string {
	return fmt.Sprintf(
		`
resource "maas_vm_host_machine" "test" {
  vm_host  = %q
  cores    = 1
  memory   = 4096  # set to above the default
  hostname = %q
}

`, vmHost, hostname)
}

func testAccMAASInstanceConfigBasic(vmHost, hostname string) string {
	return fmt.Sprintf(`
%s

resource "maas_instance" "test" {
  allocate_params {
    hostname      = maas_vm_host_machine.test.hostname
    min_memory    = 4000
    min_cpu_count = 1
  }
}
`, testAccMAASInstanceConfigSetup(vmHost, hostname))
}

func testAccMAASInstanceConfigReleaseScripts() string {
	return `
resource "maas_instance" "test" {
  release_params {
	scripts    = [maas_node_script.dummy_release_script.name]
  }

  allocate_params {
    hostname      = maas_vm_host_machine.test.hostname
    min_memory    = 4000
    min_cpu_count = 1
  }

}
`
}

func testAccMAASInstanceConfigReleaseScriptSetup(scriptName string) string {
	return fmt.Sprintf(
		`

resource "maas_node_script" "dummy_release_script" {
  script = base64encode(<<-EOF
#!/usr/bin/bash
#
# --- Start MAAS 1.0 script metadata ---
# name: %s
# title: Terraform Dummy Release Script
# script_type: release
# --- End MAAS 1.0 script metadata ---
echo "Goodbye world!"
EOF
  )
}

`, scriptName)
}

func testAccMAASInstanceConfigReleaseParams(comment, erase, force, quickErase, secureErase string) string {
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
`, comment, erase, force, quickErase, secureErase)
}

func testAccMAASInstanceConfigAllocateParamsNonDefaultArchitecture(vmHost, hostname, architecture string) string {
	return fmt.Sprintf(`
%s

resource "maas_instance" "test" {
  allocate_params {
    hostname      = maas_vm_host_machine.test.hostname
    min_memory    = 4000
    min_cpu_count = 1
    architecture  = %q
  }
}
`, testAccMAASInstanceConfigSetup(vmHost, hostname), architecture)
}
