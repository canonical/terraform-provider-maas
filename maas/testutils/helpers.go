package testutils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"strings"
	"time"
)

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
