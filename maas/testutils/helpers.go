package testutils

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// RandomMAC generates a random locally administered MAC address.
func RandomMAC() string {
	mac := make([]byte, 6)

	// Fill the slice with random bytes
	rand.Read(mac)

	// Ensure the MAC address is valid:
	// - Bit 0 of the first byte is cleared (ensuring it's a unicast address)
	// - Bit 1 of the first byte is set (marking it as locally administered)
	mac[0] = (mac[0] & 0xFE) | 0x02

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// Generate a random CIDR of the form 10.x.y.0/24, where x and y are random numbers in the usable range of 50 to 255
func GenerateRandomCidr() string {
	// Create and log a seed if required for test reproducibility
	seed := time.Now().UnixNano()
	rand.New(rand.NewSource(seed))
	log.Printf("Seed used for random CIDR: %d", seed)

	// Arbitrary minIpRange to ensure the CIDR generated doesn't conflict with any existing MAAS networks
	minIpRange := 50
	maxIpRange := 255

	cidr := fmt.Sprintf("10.%d.%d.0/24", generateRandomNumberInRange(minIpRange, maxIpRange), generateRandomNumberInRange(minIpRange, maxIpRange))
	return cidr
}

// Returns the network prefix from a CIDR. For example 10.77.77.0/24 would return 10.77.77
func GetNetworkPrefixFromCidr(cidr string) string {
	return strings.Join(strings.Split(cidr, ".")[:3], ".")
}

func generateRandomNumberInRange(min int, max int) int {
	return rand.Intn(max-min) + min
}
