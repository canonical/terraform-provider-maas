package testutils

import (
	"crypto/rand"
	"fmt"
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
