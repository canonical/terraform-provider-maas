package maas

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
)

// commissioningOutputScript is the commissioning script whose stdout MAAS
// processes into the machine's hardware inventory: the machine-resources JSON.
const commissioningOutputScript = "50-maas-01-commissioning"

// commissioningDisk is the subset of a machine-resources storage disk that the
// block_devices attribute of data.maas_machine exposes. None of these fields
// are stored on MAAS's block device model.
type commissioningDisk struct {
	Name       string `json:"id"`
	DevicePath string `json:"device_path"`
	PCIAddress string `json:"pci_address"`
	DeviceID   string `json:"device_id"`
	NUMANode   int    `json:"numa_node"`
}

type commissioningResources struct {
	Resources struct {
		Storage struct {
			Disks []commissioningDisk `json:"disks"`
		} `json:"storage"`
	} `json:"resources"`
}

// getCurrentCommissioningDisks returns the disks reported by the machine's
// current commissioning run, keyed by the id_path MAAS stores for each block
// device ("/dev/disk/by-id/<device_id>").
//
// It reads the "current-commissioning" result alias rather than listing
// /nodes/<id>/results/: the listing returns every historic script set with its
// output inlined, which is several megabytes per machine, in no guaranteed
// order, and older sets carry kernel device names that no longer apply.
//
// A machine that has never been commissioned yields an empty map, not an error.
func getCurrentCommissioningDisks(c *client.Client, systemID string) (map[string]commissioningDisk, error) {
	// gomaasclient has no call for a single script result or its download
	// operation, but every endpoint client carries the generic APIClient.
	nodeResults, ok := c.NodeResults.(*client.NodeResults)
	if !ok {
		return nil, fmt.Errorf("unexpected MAAS client type %T for node results", c.NodeResults)
	}

	params := url.Values{}
	params.Set("filters", commissioningOutputScript)
	params.Set("output", "stdout")

	var output []byte

	err := nodeResults.APIClient.
		GetSubObject("nodes").
		GetSubObject(systemID).
		GetSubObject("results").
		GetSubObject("current-commissioning").
		Get("download", params, func(data []byte) error {
			output = data
			return nil
		})
	if err != nil {
		if isNotFoundError(err) {
			return map[string]commissioningDisk{}, nil
		}

		return nil, fmt.Errorf("reading current commissioning output of machine %s: %w", systemID, err)
	}

	disks, err := parseCommissioningDisks(output)
	if err != nil {
		return nil, fmt.Errorf("parsing current commissioning output of machine %s: %w", systemID, err)
	}

	return disks, nil
}

// getDataSourceBlockDeviceParameters flattens block devices for
// data.maas_machine: the attributes shared with the maas_machine resource, plus
// device_path, pci_address and numa_node, which only the data source exposes.
//
// disks is the current commissioning output keyed by id_path (see
// getCurrentCommissioningDisks). The join is on id_path, never on name: the
// name is what a storage configuration renames, while id_path is stable. A
// device with no match keeps device_path and pci_address empty and takes
// numa_node from the value MAAS stores.
func getDataSourceBlockDeviceParameters(blockDevices []entity.BlockDevice, disks map[string]commissioningDisk) []map[string]any {
	params := getShortBlockDeviceMachineParameters(blockDevices)

	// getShortBlockDeviceMachineParameters reorders its input, so match the
	// flattened entries back to their devices by ID rather than by position.
	byID := make(map[int]entity.BlockDevice, len(blockDevices))
	for _, blockDevice := range blockDevices {
		byID[blockDevice.ID] = blockDevice
	}

	for _, p := range params {
		blockDevice := byID[p["id"].(int)]

		p["device_path"] = ""
		p["pci_address"] = ""
		p["numa_node"] = blockDevice.NUMANode

		if disk, ok := disks[blockDevice.IDPath]; ok {
			p["device_path"] = disk.DevicePath
			p["pci_address"] = disk.PCIAddress
			p["numa_node"] = disk.NUMANode
		}
	}

	return params
}

// parseCommissioningDisks extracts the storage disks from machine-resources
// JSON, keyed by "/dev/disk/by-id/<device_id>". That is the id_path MAAS
// records for a physical block device, and unlike the device name it does not
// change when the device is renamed. Disks without a device_id are skipped,
// since nothing can join them to a block device reliably.
func parseCommissioningDisks(output []byte) (map[string]commissioningDisk, error) {
	disks := map[string]commissioningDisk{}

	if len(output) == 0 {
		return disks, nil
	}

	var resources commissioningResources
	if err := json.Unmarshal(output, &resources); err != nil {
		return nil, err
	}

	for _, disk := range resources.Resources.Storage.Disks {
		if disk.DeviceID == "" {
			continue
		}

		disks["/dev/disk/by-id/"+disk.DeviceID] = disk
	}

	return disks, nil
}
