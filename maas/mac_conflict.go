package maas

import (
	"errors"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
)

// errMachineNotFound is returned by findMachineByMAC when no machine matches
// the supplied MAC. Callers should use errors.Is to detect it instead of
// branching on a nil-machine value.
var errMachineNotFound = errors.New("maas: machine not found")

// isMACConflict reports whether err is the MAAS 400 response that fires when a
// machine with the same MAC address already exists. MAAS does not expose a
// machine-readable error code; gomaasapi formats the response as
//
//	ServerError: 400 Bad Request (<json or list body>)
//
// Body shapes observed in practice (across MAAS 2.x / 3.x):
//   - {"mac_addresses": ["MAC address xx:xx... already in use on <host>."]}
//   - {"__all__":       ["A node with this MAC address already exists."]}
//   - ["MAC address xx:xx... already in use on <host>."]
//
// Match strategy: status 400 + "mac" (case-insensitive) + "already" or "exists".
// Conservative on purpose: avoids matching unrelated 400s, accepts any of the
// observed phrasings.
func isMACConflict(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "400") {
		return false
	}

	if !strings.Contains(msg, "mac") {
		return false
	}

	return strings.Contains(msg, "already") || strings.Contains(msg, "exists")
}

// findMachineByMAC returns the MAAS machine whose boot-interface MAC matches
// mac. Returns errMachineNotFound when no machine matches; callers should use
// errors.Is(err, errMachineNotFound) to distinguish from API errors.
func findMachineByMAC(c *client.Client, mac string) (*entity.Machine, error) {
	machines, err := c.Machines.Get(&entity.MachinesParams{
		MACAddress: []string{mac},
	})
	if err != nil {
		return nil, err
	}

	want := strings.ToLower(strings.TrimSpace(mac))
	for i := range machines {
		got := strings.ToLower(strings.TrimSpace(machines[i].BootInterface.MACAddress))
		if got == want {
			return &machines[i], nil
		}
	}

	return nil, errMachineNotFound
}
