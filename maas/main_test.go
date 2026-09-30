package maas_test

import (
	"log"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	sweepLeftoverInterfaces()
	resource.TestMain(m)
}

// sweepLeftoverInterfaces deletes test interfaces that an earlier process left
// behind on the acceptance machine. Generated names are only unique within a
// process, so anything surviving a crashed run would collide with this one.
// Failures are logged rather than fatal: the acceptance tests skip themselves
// when the MAAS_* environment is missing anyway.
func sweepLeftoverInterfaces() {
	if slices.Contains(os.Args, "-sweep") {
		// The registered sweeper covers this and reports its own failures.
		return
	}

	machine := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	if machine == "" {
		return
	}

	clientConfig, err := getSweeperClient()
	if err != nil {
		log.Printf("[WARN] Skipping pre-suite interface sweep: %s", err)
		return
	}

	interfaces, err := clientConfig.Client.NetworkInterfaces.Get(machine)
	if err != nil {
		log.Printf("[WARN] Skipping pre-suite interface sweep: %s", err)
		return
	}

	for _, iface := range interfaces {
		if !strings.HasPrefix(iface.Name, "tf-nic-") {
			continue
		}

		log.Printf("[INFO] Pre-suite sweep: deleting leftover interface %s (ID: %d)", iface.Name, iface.ID)

		if err := clientConfig.Client.NetworkInterface.Delete(machine, iface.ID); err != nil {
			log.Printf("[WARN] Pre-suite sweep: could not delete interface %s: %s", iface.Name, err)
		}
	}
}
