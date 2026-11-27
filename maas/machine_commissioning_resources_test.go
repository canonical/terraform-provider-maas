package maas

import (
	"os"
	"testing"

	"github.com/canonical/gomaasclient/entity"
)

// testdata/commissioning_resources.json is the storage section of a real
// 50-maas-01-commissioning output: 2 single-namespace OS drives and 4 Ceph
// drives split into 3 namespaces each, plus a synthetic CD-ROM with no
// device_id. Kernel names deliberately do not follow PCI order there.
func loadCommissioningFixture(t *testing.T) map[string]commissioningDisk {
	t.Helper()

	data, err := os.ReadFile("testdata/commissioning_resources.json")
	if err != nil {
		t.Fatal(err)
	}

	disks, err := parseCommissioningDisks(data)
	if err != nil {
		t.Fatal(err)
	}

	return disks
}

func TestParseCommissioningDisks(t *testing.T) {
	disks := loadCommissioningFixture(t)

	if got, want := len(disks), 14; got != want {
		t.Fatalf("got %d disks, want %d (the disk without a device_id must be skipped)", got, want)
	}

	for idPath, disk := range disks {
		if idPath != "/dev/disk/by-id/"+disk.DeviceID {
			t.Errorf("disk %s keyed as %q, want /dev/disk/by-id/%s", disk.Name, idPath, disk.DeviceID)
		}
	}

	disk, ok := disks["/dev/disk/by-id/nvme-eui.00000000000000030000000000000003"]
	if !ok {
		t.Fatal("namespace 2 of the 0000:03:00.0 controller is missing")
	}

	want := commissioningDisk{
		Name:       "nvme2n1",
		DevicePath: "pci-0000:03:00.0-nvme-2",
		PCIAddress: "0000:03:00.0",
		DeviceID:   "nvme-eui.00000000000000030000000000000003",
		NUMANode:   0,
	}
	if disk != want {
		t.Errorf("got %+v, want %+v", disk, want)
	}
}

func TestParseCommissioningDisksEmpty(t *testing.T) {
	for name, input := range map[string][]byte{
		"empty body":     nil,
		"no resources":   []byte(`{}`),
		"no disks":       []byte(`{"resources": {"storage": {}}}`),
		"null disk list": []byte(`{"resources": {"storage": {"disks": null}}}`),
	} {
		t.Run(name, func(t *testing.T) {
			disks, err := parseCommissioningDisks(input)
			if err != nil {
				t.Fatal(err)
			}

			if len(disks) != 0 {
				t.Errorf("got %d disks, want none", len(disks))
			}
		})
	}
}

func TestParseCommissioningDisksInvalid(t *testing.T) {
	if _, err := parseCommissioningDisks([]byte("not json")); err == nil {
		t.Error("expected an error for non-JSON output")
	}
}

func TestDataSourceBlockDeviceParametersCommissioningJoin(t *testing.T) {
	disks := loadCommissioningFixture(t)

	blockDevices := []entity.BlockDevice{
		// Renamed in MAAS since commissioning: the join must still hold.
		{ID: 2, Name: "nvmef1n1", IDPath: "/dev/disk/by-id/nvme-eui.00000000000000010000000000000004", NUMANode: 0},
		{ID: 1, Name: "nvme2n1", IDPath: "/dev/disk/by-id/nvme-eui.00000000000000030000000000000003", NUMANode: 0},
		// Not in the commissioning output.
		{ID: 3, Name: "sda", IDPath: "/dev/disk/by-id/scsi-SQEMU_QEMU_HARDDISK_lxd_root", NUMANode: 1},
	}

	params := getDataSourceBlockDeviceParameters(blockDevices, disks)

	want := []map[string]any{
		{"name": "nvme2n1", "device_path": "pci-0000:03:00.0-nvme-2", "pci_address": "0000:03:00.0", "numa_node": 0},
		{"name": "nvmef1n1", "device_path": "pci-0000:f1:00.0-nvme-1", "pci_address": "0000:f1:00.0", "numa_node": 1},
		{"name": "sda", "device_path": "", "pci_address": "", "numa_node": 1},
	}

	if len(params) != len(want) {
		t.Fatalf("got %d block devices, want %d", len(params), len(want))
	}

	for i, w := range want {
		for key, value := range w {
			if params[i][key] != value {
				t.Errorf("block device %d (%v): %s = %v, want %v", i, params[i]["name"], key, params[i][key], value)
			}
		}
	}
}

func TestDataSourceBlockDeviceParametersWithoutCommissioning(t *testing.T) {
	params := getDataSourceBlockDeviceParameters([]entity.BlockDevice{
		{ID: 1, Name: "nvme2n1", IDPath: "/dev/disk/by-id/nvme-eui.00000000000000030000000000000003", NUMANode: 1},
	}, nil)

	if params[0]["device_path"] != "" || params[0]["pci_address"] != "" || params[0]["numa_node"] != 1 {
		t.Errorf("without commissioning output: got %v", params[0])
	}
}
