package maas_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"terraform-provider-maas/maas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("maas_network_interface", &resource.Sweeper{
		Name: "maas_network_interface",
		F:    sweepNetworkInterfaces,
	})
}

func sweepNetworkInterfaces(region string) error {
	// Get the test machine system ID from environment
	machineSystemID := os.Getenv("TF_ACC_NETWORK_INTERFACE_MACHINE")
	if machineSystemID == "" {
		log.Printf("[INFO] TF_ACC_NETWORK_INTERFACE_MACHINE not set, skipping network interface sweep")
		return nil
	}

	log.Printf("[INFO] Starting network interface sweep for machine: %s", machineSystemID)

	// Get MAAS client
	clientConfig, err := getSweeperClient()
	if err != nil {
		return fmt.Errorf("error getting sweeper client: %s", err)
	}

	// Get all network interfaces for the machine
	interfaces, err := clientConfig.Client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return fmt.Errorf("error getting network interfaces for machine %s: %s", machineSystemID, err)
	}

	log.Printf("[INFO] Found %d total interfaces on machine %s", len(interfaces), machineSystemID)

	// Find and delete all test interfaces (created by tests)
	// This handles physical, vlan, bridge, and bond interfaces
	deletedCount := 0

	for _, iface := range interfaces {
		// Only delete interfaces with test naming pattern
		if !strings.HasPrefix(iface.Name, "tf-nic-") {
			log.Printf("[DEBUG] Skipping interface: %s (does not match tf-nic- pattern)", iface.Name)
			continue
		}

		log.Printf("[INFO] Deleting test interface: %s (Type: %s, ID: %d, MAC: %s)", iface.Name, iface.Type, iface.ID, iface.MACAddress)

		// Delete the interface
		if err := clientConfig.Client.NetworkInterface.Delete(machineSystemID, iface.ID); err != nil {
			log.Printf("[ERROR] Failed to delete interface %s: %s", iface.Name, err)
			// Continue trying to delete others
			continue
		}

		deletedCount++

		log.Printf("[INFO] Successfully deleted interface: %s", iface.Name)
	}

	log.Printf("[INFO] Network interface sweep complete: deleted %d interfaces", deletedCount)

	return nil
}

// getSweeperClient creates a MAAS client for use in sweepers
func getSweeperClient() (*maas.ClientConfig, error) {
	apiKey := os.Getenv("MAAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MAAS_API_KEY must be set for sweepers")
	}

	apiURL := os.Getenv("MAAS_API_URL")
	if apiURL == "" {
		return nil, fmt.Errorf("MAAS_API_URL must be set for sweepers")
	}

	config := maas.Config{
		APIKey:     apiKey,
		APIURL:     apiURL,
		APIVersion: "2.0",
	}

	client, err := config.Client()
	if err != nil {
		return nil, fmt.Errorf("error creating MAAS client: %s", err)
	}

	version, err := client.Version.Get()
	if err != nil {
		return nil, fmt.Errorf("error getting MAAS version: %s", err)
	}

	return &maas.ClientConfig{
		Client:      client,
		MAASVersion: version.Version,
	}, nil
}
