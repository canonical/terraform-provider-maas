package maas_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMAASVMHost_DeployParams(t *testing.T) {
	// A VM host identifier. Used to create a VM, which is deployed as a VM host.
	vmHostIdentifier := os.Getenv("TF_ACC_VM_HOST_ID")

	testMachineName := acctest.RandomWithPrefix("tf-resource-vm-host")
	vmHostType := "lxd"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_VM_HOST_ID"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASVMHostDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASVMHostDeployParamsConfig(vmHostIdentifier, testMachineName, vmHostType),
				Check: resource.ComposeTestCheckFunc(
					checkMAASVMHostExists(t, "maas_vm_host.test"),
					resource.TestCheckResourceAttr("maas_vm_host.test", "type", "lxd"),
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

func testAccMAASVMHostDeployParamsConfig(vmHostIdentifier string, testMachineName string, vmHostType string) string {
	return fmt.Sprintf(`
	resource "maas_vm_host_machine" "test" {
	  vm_host  = %q
	  hostname = %q
	}

	resource "maas_vm_host" "test" {
	  machine = maas_vm_host_machine.test.id
	  type    = %q

	  deploy_params {
		  user_data        = "#!/bin/bash\necho 'Hello from cloud-init'"
	  }
	}
	`, vmHostIdentifier, testMachineName, vmHostType)
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
