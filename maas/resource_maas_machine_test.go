package maas_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-maas/maas/testutils"

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
func testAccDataSourceMAASMachineLookup(hostname string) string {
	return fmt.Sprintf(`
data "maas_machine" "test" {
  hostname = "%s"
}

resource "maas_network_interface_physical" "test_nic1" {
  machine     = data.maas_machine.test.id
  mac_address = "52:54:00:7c:f7:77"
  name        = "tf-nic-lookup-by-id"
}

resource "maas_network_interface_physical" "test_nic2" {
  machine     = "%s"
  mac_address = "52:54:00:7c:f7:78"
  name        = "tf-nic-lookup-by-hostname"
}
`, hostname, hostname)
}

// TestAccResourceMAASMachine_Lookup verifies that an existing MAAS machine
// can be resolved correctly using both its system_id and its hostname.
// The machine hostname is supplied via TF_ACC_MACHINE_HOSTNAME.
func TestAccResourceMAASMachine_Lookup(t *testing.T) {
	hostname := os.Getenv("TF_ACC_MACHINE_HOSTNAME")

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
				Config: testAccDataSourceMAASMachineLookup(hostname),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}
