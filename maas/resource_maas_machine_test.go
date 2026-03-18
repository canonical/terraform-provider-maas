package maas_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"terraform-provider-maas/maas/testutils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testAccDataSourceMAASMachineLookup generates a Terraform configuration that:
//
// 1. Looks up an existing MAAS machine via the data source using its hostname.
// 2. Creates two physical network interfaces on that machine:
//   - One referencing the machine by its system_id
//   - One referencing the machine by its hostname
//
// This setup intentionally exercises both code paths inside the provider's
// getMachine() helper:
//
//   - getMachine(client, system_id)   -> direct lookup via Machine.Get()
//   - getMachine(client, hostname)    -> fallback lookup via Machines.Get()
//
// Terraform does not allow passing system_id directly into the maas_machine
// data source, so we validate lookup-by-ID indirectly by passing the ID into
// another resource that internally resolves the machine using getMachine().
func testAccDataSourceMAASMachineLookup(hostname string, macAddress1 string, nicName1 string, macAddress2 string, nicName2 string) string {
	return fmt.Sprintf(`
data "maas_machine" "test" {
  hostname = "%s"
}

resource "maas_network_interface_physical" "test_nic1" {
  machine     = data.maas_machine.test.id
  mac_address = "%s"
  name        = "%s"
}

resource "maas_network_interface_physical" "test_nic2" {
  machine     = "%s"
  mac_address = "%s"
  name        = "%s"
}
`, hostname, macAddress1, nicName1, hostname, macAddress2, nicName2)
}

// TestAccResourceMAASMachine_Lookup verifies that an existing MAAS machine
// can be resolved correctly using both its system_id and its hostname.
// The machine hostname is supplied via TF_ACC_MACHINE_HOSTNAME.
func TestAccResourceMAASMachine_Lookup(t *testing.T) {
	nicName1 := fmt.Sprintf("tf-lookup-1-%d", acctest.RandIntRange(0, 9))
	nicName2 := fmt.Sprintf("tf-lookup-2-%d", acctest.RandIntRange(0, 9))
	hostname := os.Getenv("TF_ACC_MACHINE_HOSTNAME")
	macAddress1 := testutils.RandomMAC()
	macAddress2 := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.maas_machine.test", "hostname", hostname),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_MACHINE_HOSTNAME"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMAASMachineLookup(hostname, macAddress1, nicName1, macAddress2, nicName2),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccResourceMAASMachine_NoPXE(t *testing.T) {
	testLXDIP := "10.0.0.10"
	// This could pull from TF args for the actual BMC address of a machine and remove the PlanOnly below.
	testIPMIIP := "10.0.0.10"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config:      testAccMAASMachineLXDNoPXE(testLXDIP),
				ExpectError: regexp.MustCompile(`pxe_mac_address is required when power_type is not 'ipmi'`),
			},
			{
				Config: testAccMAASMachineIPMINoPXE(testIPMIIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("maas_machine.test", "power_type", "ipmi"),
				),
				// Verify the plan is valid, don't actually create the machines.
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMAASMachineLXDNoPXE(ipAddress string) string {
	return fmt.Sprintf(`
resource "maas_machine" "test" {
  power_type = "lxd"
  power_parameters = jsonencode({
    power_address = %q
    instance_name = "test"
  })
  hostname = "lxdTestMachineNoPxe"
}
`, ipAddress)
}

func testAccMAASMachineIPMINoPXE(ipAddress string) string {
	return fmt.Sprintf(`
resource "maas_machine" "test" {
  power_type = "ipmi"
  architecture = "amd64/generic"
  power_parameters = jsonencode({
    power_address = %q
    power_user    = "admin"
    power_pass    = "password"
  })
  hostname = "ipmiTestMachineNoPxe"
}
`, ipAddress)
}
