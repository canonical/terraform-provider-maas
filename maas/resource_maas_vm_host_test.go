package maas_test

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMAASVMHost_basic(t *testing.T) {
	VMHostName := acctest.RandomWithPrefix("tf-vm-host")

	var powerAddress string

	project := "mass"
	VMHostID := os.Getenv("TF_ACC_VM_HOST_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASVMHostDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			// Test create
			{
				PreConfig: func() {
					client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
					id, err := strconv.Atoi(VMHostID)
					if err != nil {
						panic(err)
					}
					host, err := client.VMHost.GetParameters(id)
					if err != nil {
						panic(fmt.Errorf("Error getting VM Host (%v): %v", VMHostID, err))
					}
					existingPowerAddr := host["power_address"]

					powerAddress = existingPowerAddr
				},
				Config: testAccMAASVMHostConfig(VMHostName, project, powerAddress),
				Check: resource.ComposeTestCheckFunc(
					checkMAASVMHostExists(t, "maas_vm_host.test"),
					resource.TestCheckResourceAttr("maas_vm_host.test", "type", "lxd"),
					resource.TestCheckResourceAttr("maas_vm_host.test", "power_address", powerAddress),
					resource.TestCheckResourceAttr("maas_vm_host.test", "project", project),
					checkMAASVMHostIsComposable(t, "maas_vm_host.test"),
				),
			},
		},
	})
}

func testAccMAASVMHostConfig(powerAddress, projectName, vmHostName string) string {
	return fmt.Sprintf(`
resource "maas_vm_host" "test" {
  type          = "lxd"
  power_address = %q
  project       = %q
  certificate   = file(systemtests.cert)
  key           = file(systemtests.key)
  name          = %q
}
	`, powerAddress, projectName, vmHostName)
}

func checkMAASVMHostIsComposable(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		systemID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to convert system ID to int: %s", err)
		}

		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		machine, err := client.VMHost.Get(systemID)
		if err != nil {
			return err
		}

		if !slices.Contains(machine.Capabilities, "composable") {
			return fmt.Errorf("VM host  is not composable")
		}

		return nil
	}
}
func TestAccMAASVMHost_DeployParams(t *testing.T) {
	// A VM host identifier. Used to create a VM, which is deployed as a VM host in this test.
	vmHostIdentifier := os.Getenv("TF_ACC_VM_HOST_ID")
	// A random string to be used for test
	rs := acctest.RandString(8)
	testMachineName := fmt.Sprintf("test-vm-host-machine-%s", rs)
	testVMHostName := fmt.Sprintf("test-vm-host-%s", rs)
	resourceName := fmt.Sprintf("maas_vm_host.%s", testVMHostName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASVMHostDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASVMHostDeployParamsConfig(vmHostIdentifier, testMachineName, testVMHostName),
				Check: resource.ComposeTestCheckFunc(
					checkMAASVMHostExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "lxd"),
				),
			},
		},
	})
}

func checkMAASVMHostExists(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Log("Checking if VM host exists...")

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		systemID := rs.Primary.Attributes["machine"]
		client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		var defaultDistroSeries string

		defaultDistroSeriesbytes, err := client.MAASServer.Get("default_distro_series")
		if err != nil {
			t.Fatalf("Failed to get default Distro Series from client: %s", err)
		}

		err = json.Unmarshal(defaultDistroSeriesbytes, &defaultDistroSeries)
		if err != nil {
			t.Fatalf("Failed to unmarshal defaultDistroSeriesbytes: %s", err)
		}

		machine, err := client.Machine.Get(systemID)
		if err != nil {
			return err
		}

		if machine.DistroSeries != defaultDistroSeries {
			return fmt.Errorf("Distro series not the expected default: %s, expected: %s", machine.DistroSeries, defaultDistroSeries)
		}

		t.Log("VM host exists!")

		return nil
	}
}

func testAccMAASVMHostDeployParamsConfig(vmHostIdentifier string, testMachineName string, testVMHostName string) string {
	return fmt.Sprintf(`
	resource "maas_vm_host_machine" "%s" {
	  vm_host = %q
	  cores   = 1
	  memory  = 2048

	  storage_disks {
	    size_gigabytes = 12
	  }
	}
	resource "maas_vm_host" "%s" {
	  machine = maas_vm_host_machine.%s.id
	  type    = "lxd"

	  deploy_params {
		  user_data        = "#!/bin/bash\necho 'Hello from cloud-init'"
	  }
	}
	`, testMachineName, vmHostIdentifier, testVMHostName, testMachineName)
}

func testAccCheckMAASVMHostDestroy(s *terraform.State) error {
	client := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_vm_host" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := client.VMHost.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("VM host still exists")
			}
		}

		// If the error is a 404 not found error, the VM host is destroyed
		if err != nil && strings.Contains(err.Error(), "404 Not Found") {
			continue
		}

		return err
	}

	return nil
}
