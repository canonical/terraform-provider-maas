package testutils

import (
	"strconv"
	"strings"
	"testing"
)

func TestGenerateRandomNumberInRange(t *testing.T) {
	min := 0
	max := 5
	iterations := 200

	for range iterations {
		result := generateRandomNumberInRange(min, max)
		if result < min || result > max {
			t.Errorf("Generated number %d is less than minimum %d or greater than maximum %d", result, min, max)
		}
	}
}

func TestGetNetworkPrefixFromCIDR(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"192.168.1.0/24", "192.168.1"},
		{"10.0.0.0/8", "10.0.0"},
	}

	for _, test := range tests {
		prefix := GetNetworkPrefixFromCIDR(test.input)
		if prefix != test.output {
			t.Errorf("Prefix should be %s, got %s", test.output, prefix)
		}
	}
}

func TestGenerateRandomCIDR(t *testing.T) {
	cidr := GenerateRandomCIDR()

	parts := strings.Split(cidr, ".")
	if len(parts) != 4 {
		t.Errorf("CIDR should have 4 parts, got %d parts: %s", len(parts), cidr)
	}
	// Check the first and last octets
	if parts[0] != "10" {
		t.Errorf("First octet should be 10, got %s", parts[0])
	}

	lastPart := parts[3]

	if lastPart != "0/24" {
		t.Errorf("CIDR should end with 0/24, got %s", lastPart)
	}
	// Check middle octets are within the range of 0 to 255
	for i := 1; i <= 2; i++ {
		octet, err := strconv.Atoi(parts[i])
		if err != nil {
			t.Errorf("Failed to convert octet to integer: %v", err)
		}

		if octet < 50 || octet > 255 {
			t.Errorf("Octet %d should be between 50 and 255, got %d", i, octet)
		}
	}
}
