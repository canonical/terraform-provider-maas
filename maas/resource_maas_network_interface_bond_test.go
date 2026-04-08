package maas_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccMAASNetworkInterfaceSetup(machine string, fabricName string, physOneName string, physOneMAC string, physTwoName string, physTwoMAC string) string {
	return fmt.Sprintf(`
resource "maas_fabric" "default" {
	name = "tf-fabric-%s"
}

data "maas_machine" "machine" {
	hostname = "%s"
}

data "maas_vlan" "default" {
	fabric = maas_fabric.default.id
	vlan   = 0
}

resource "maas_network_interface_physical" "nic1" {
	machine     = data.maas_machine.machine.id
	name        = "%s"
	mac_address = "%s"
	vlan        = data.maas_vlan.default.id
}

resource "maas_network_interface_physical" "nic2" {
	machine     = data.maas_machine.machine.id
	name        = "%s"
	mac_address = "%s"
	vlan        = data.maas_vlan.default.id
}`, fabricName, machine, physOneName, physOneMAC, physTwoName, physTwoMAC)
}

func testAccMAASNetworkInterfaceBond(name string, machine string, fabric string, physOneName string, physTwoName string, macAddress string, macAddressPhysOne string, macAddressPhysTwo string, mtu int) string {
	setup := testAccMAASNetworkInterfaceSetup(machine, fabric, physOneName, macAddressPhysOne, physTwoName, macAddressPhysTwo)

	return fmt.Sprintf(`
%v

resource "maas_network_interface_bond" "test" {
	machine               = data.maas_machine.machine.id
	name                  = "%s"
	accept_ra             = false
	bond_downdelay        = 1
	bond_lacp_rate        = "slow"
	bond_miimon           = 10
	bond_mode             = "802.3ad"
	bond_num_grat_arp     = 1
	bond_updelay          = 1
	bond_xmit_hash_policy = "layer2"
	mac_address           = "%s"
	mtu                   = %d
	parents               = [maas_network_interface_physical.nic1.name, maas_network_interface_physical.nic2.name]
	tags                  = ["tag1", "tag2"]
	vlan                  = data.maas_vlan.default.id
}
`, setup, name, macAddress, mtu)
}

func TestAccResourceMAASNetworkInterfaceBond_basic(t *testing.T) {
	var networkInterfaceBond entity.NetworkInterface

	name := fmt.Sprintf("tf-nic-bond-%d", acctest.RandIntRange(0, 9))
	fabric := "bond"
	physOne := "enp109s0f0"
	physTwo := "enp109s0f1"
	machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	macAddress := testutils.RandomMAC()
	macAddressPhysOne := testutils.RandomMAC()
	macAddressPhysTwo := testutils.RandomMAC()

	checks := []resource.TestCheckFunc{
		testAccMAASNetworkInterfaceBondCheckExists("maas_network_interface_bond.test", &networkInterfaceBond),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "name", name),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "accept_ra", "false"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_downdelay", "1"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_lacp_rate", "slow"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_miimon", "10"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_mode", "802.3ad"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_num_grat_arp", "1"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_updelay", "1"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "bond_xmit_hash_policy", "layer2"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "mac_address", macAddress),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "parents.#", "2"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "parents.0", physOne),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "parents.1", physTwo),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "tags.#", "2"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "tags.0", "tag1"),
		resource.TestCheckResourceAttr("maas_network_interface_bond.test", "tags.1", "tag2"),
		resource.TestCheckResourceAttrPair("maas_network_interface_bond.test", "vlan", "data.maas_vlan.default", "id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_NETWORK_INTERFACE_MACHINE"}) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASNetworkInterfaceBondDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASNetworkInterfaceBond(name, machine, fabric, physOne, physTwo, macAddress, macAddressPhysOne, macAddressPhysTwo, 1500),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_bond.test", "mtu", "1500"))...),
			},
			// Test update
			{
				Config: testAccMAASNetworkInterfaceBond(name, machine, fabric, physOne, physTwo, macAddress, macAddressPhysOne, macAddressPhysTwo, 9000),
				Check: resource.ComposeTestCheckFunc(
					append(checks, resource.TestCheckResourceAttr("maas_network_interface_bond.test", "mtu", "9000"))...),
			},
			// Test import
			{
				ResourceName:      "maas_network_interface_bond.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["maas_network_interface_bond.test"]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", "maas_network_interface_bond.test")
					}

					if rs.Primary.ID == "" {
						return "", fmt.Errorf("resource id not set")
					}

					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["machine"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func TestAccResourceMAASNetworkInterfaceBond_404Handling(t *testing.T) {
    machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
    if machine == "" {
        t.Fatal("TF_ACC_NETWORK_INTERFACE_MACHINE must be set")
    }

    resourceName := "maas_network_interface_bond.test"

    testMatrix := []struct {
        name           string
        DeleteFunction func(client *client.Client, systemID string, bondID int) error
        testUpdate     bool
    }{
        {
            name: "Update bond when missing triggers error",
            DeleteFunction: func(client *client.Client, systemID string, bondID int) error {
                return client.NetworkInterface.Delete(systemID, bondID)
            },
            testUpdate: true,
        },
        {
            name: "Update bond with missing parent triggers error",
            DeleteFunction: func(client *client.Client, systemID string, bondID int) error {
                return deleteBondParent(client, systemID, bondID)
            },
            testUpdate: true,
        },
        {
            name: "Delete bond when missing triggers error",
            DeleteFunction: func(client *client.Client, systemID string, bondID int) error {
                return client.NetworkInterface.Delete(systemID, bondID)
            },
            testUpdate: false,
        },
        {
            name: "Delete bond with missing parent triggers error",
            DeleteFunction: func(client *client.Client, systemID string, bondID int) error {
                return deleteBondParent(client, systemID, bondID)
            },
            testUpdate: false,
        },
    }

    for _, thisTest := range testMatrix {
        t.Run(thisTest.name, func(t *testing.T) {
            var networkInterfaceBond entity.NetworkInterface
            var systemID string

            name := fmt.Sprintf("tf-bond-%s", acctest.RandString(4))
            fabric := fmt.Sprintf("bond-%s", acctest.RandString(10))
            physOne := fmt.Sprintf("enp109s0f-%s", acctest.RandString(4))
            physTwo := fmt.Sprintf("enp109s0f-%s", acctest.RandString(4))

            macAddress := testutils.RandomMAC()
            macAddressPhysOne := testutils.RandomMAC()
            macAddressPhysTwo := testutils.RandomMAC()

            resource.ParallelTest(t, resource.TestCase{
                PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_NETWORK_INTERFACE_MACHINE"}) },
                Providers:    testutils.TestAccProviders,
                CheckDestroy: testAccCheckMAASNetworkInterfaceBondDestroy,
                ErrorCheck:   func(err error) error { return err },
                Steps: func() []resource.TestStep {
                    steps := []resource.TestStep{
                        {
                            Config: testAccMAASNetworkInterfaceBond(name, machine, fabric, physOne, physTwo, macAddress, macAddressPhysOne, macAddressPhysTwo, 1500),
                            Check: resource.ComposeTestCheckFunc(
                                testAccMAASNetworkInterfaceBondCheckExists(resourceName, &networkInterfaceBond),
                                func(s *terraform.State) error {
                                    rs, ok := s.RootModule().Resources[resourceName]
                                    if !ok {
                                        return fmt.Errorf("Resource not found in state: %s", resourceName)
                                    }
                                    systemID = rs.Primary.Attributes["machine"]
                                    return nil
                                },
                            ),
                        },
                    }

                    // Change the test based on delete or update
                    if thisTest.testUpdate {
                        steps = append(steps, resource.TestStep{
                            PreConfig: func() {
                                conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
                                if err := thisTest.DeleteFunction(conn, systemID, networkInterfaceBond.ID); err != nil {
                                    t.Fatalf("PreConfig failed to delete resource: %v", err)
                                }
                            },
                            Config: testAccMAASNetworkInterfaceBond(name, machine, fabric, physOne, physTwo, macAddress, macAddressPhysOne, macAddressPhysTwo, 9000),
                            ExpectError: regexp.MustCompile(`404 Not Found`),
                        })
                    } else {
                        steps = append(steps, resource.TestStep{
                            PreConfig: func() {
                                conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
                                if err := thisTest.DeleteFunction(conn, systemID, networkInterfaceBond.ID); err != nil {
                                    t.Fatalf("PreConfig failed to delete resource: %v", err)
                                }
                            },
                            Config:  " ",
                            Destroy: true,
                            ExpectError: regexp.MustCompile(`404 Not Found`),
                        })
                    }

                    return steps
                }(),
            })
        })
    }
}

func testAccMAASNetworkInterfaceBondCheckExists(rn string, networkInterfaceBond *entity.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		gotNetworkInterfaceBond, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err != nil {
			return fmt.Errorf("error getting network interface bond: %s", err)
		}

		*networkInterfaceBond = *gotNetworkInterfaceBond

		return nil
	}
}

func testAccCheckMAASNetworkInterfaceBondDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_network_interface_bond
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_network_interface_bond" {
			continue
		}

		// Retrieve our maas_network_interface_bond by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, err := conn.NetworkInterface.Get(rs.Primary.Attributes["machine"], id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Network interface bond (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_network_interface_bond is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}

func testAccCheckResourceUnset(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[rn]; ok {
			return fmt.Errorf("Resource %s still exists in state", rn)
		}

		return nil
	}
}

func getNetworkInterface(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return nil, err
	}

	for _, n := range networkInterfaces {
		if n.MACAddress == identifier || n.Name == identifier || fmt.Sprintf("%v", n.ID) == identifier {
			return &n, nil
		}
	}

	return nil, fmt.Errorf("network interface (%s) was not found on machine (%s)", identifier, machineSystemID)
}

func findBondParentsID(client *client.Client, machineSystemID string, parents []any) ([]int, error) {
	var result []int

	for _, p := range parents {
		if p, ok := p.(string); ok {
			networkInterface, err := getNetworkInterface(client, machineSystemID, p)
			if err != nil {
				return nil, err
			}

			if networkInterface.Type != "physical" {
				continue
			}

			result = append(result, networkInterface.ID)
		}
	}

	return result, nil
}

func deleteBondParent(client *client.Client, systemID string, bondID int) error {
	bond, err := client.NetworkInterface.Get(systemID, bondID)
	if err != nil {
		return err
	}
	// delete at least one parent
	if len(bond.Parents) > 0 {
		parentsAny := make([]any, len(bond.Parents))
		for i, p := range bond.Parents {
			parentsAny[i] = p
		}

		parents, err := findBondParentsID(client, systemID, parentsAny)
		if err != nil {
			return err
		}

		return client.NetworkInterface.Delete(systemID, parents[0])
	}

	return nil
}
