package testutils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"strings"
	"sync/atomic"
	"time"
)

// nicCounter hands out the suffix of generated test interface names. It is
// package state on purpose: names only have to be unique within this process.
var nicCounter atomic.Uint64

// UniqueInterfaceName builds an interface name from prefix and a process-wide
// counter. MAAS enforces interface names per node while the acceptance suite
// runs several tests against the same machine concurrently, and each step
// re-generates the config, so randomising a small range lets a later draw
// repeat a name whose interface is still present.
//
// The result is always exactly 15 characters: Linux rejects interface names
// longer than that (IFNAMSIZ minus the NUL terminator), so the counter is
// padded out to fill whatever space the prefix leaves over.
func UniqueInterfaceName(prefix string) string {
	const maxNameLen = 15

	// Truncate an over-long prefix instead of refusing to generate a name:
	// callers pass literals, and a name MAAS will reject is worse than a
	// shorter one.
	if len(prefix) >= maxNameLen {
		prefix = prefix[:maxNameLen-1]
	}

	suffixLen := maxNameLen - len(prefix)

	// Zero-pad so every generated name is the same length.
	suffix := fmt.Sprintf("%0*x", suffixLen, nicCounter.Add(1))
	if len(suffix) > suffixLen {
		// The counter has outgrown the space the prefix leaves. Keep its last
		// digits so the name still fits, rather than hand MAAS something it
		// will reject. Reaching this first takes 16**suffixLen names (4096 for
		// the 12-character prefixes in use here), which a run never generates.
		suffix = suffix[len(suffix)-suffixLen:]
	}

	return prefix + suffix
}

// RandomMAC generates a random locally administered MAC address.
func RandomMAC() string {
	mac := make([]byte, 6)

	// Fill the slice with random bytes
	_, err := rand.Read(mac)
	if err != nil {
		return "01:23:45:67:89:AB"
	}

	// Ensure the MAC address is valid:
	// - Bit 0 of the first byte is cleared (ensuring it's a unicast address)
	// - Bit 1 of the first byte is set (marking it as locally administered)
	mac[0] = (mac[0] & 0xFE) | 0x02

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// GenerateRandomCIDR generates a random CIDR of the form 10.x.y.0/24, where x and y are random numbers in the usable range of 50 to 255
func GenerateRandomCIDR() string {
	// Create and log a seed if required for test reproducibility
	seed := time.Now().UnixNano()
	mrand.New(mrand.NewSource(seed)) //nolint:gosec // used for testing only, no need for real randomness

	// Arbitrary minIPRange to ensure the CIDR generated doesn't conflict with any existing MAAS networks
	minIPRange := 50
	maxIPRange := 255

	cidr := fmt.Sprintf("10.%d.%d.0/24", generateRandomNumberInRange(minIPRange, maxIPRange), generateRandomNumberInRange(minIPRange, maxIPRange))

	return cidr
}

// GetNetworkPrefixFromCIDR returns the network prefix from a CIDR. For example 10.77.77.0/24 would return 10.77.77
func GetNetworkPrefixFromCIDR(cidr string) string {
	return strings.Join(strings.Split(cidr, ".")[:3], ".")
}

func generateRandomNumberInRange(min int, max int) int {
	return mrand.Intn(max-min) + min //nolint:gosec // used for testing only, no need for real randomness
}

// StringifySliceAsLiteralArray returns a string representation of a slice of strings, used for insertion into another string e.g., for a Terraform config.
// For example, the slice ["foo", "bar"] would become `["foo", "bar"]` where quotes and commas are actual characters in the string.
func StringifySliceAsLiteralArray(sliceOfStrings []string) string {
	sliceString, _ := json.Marshal(sliceOfStrings)
	return string(sliceString)
}
