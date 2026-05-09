package maas

import (
	"errors"
	"testing"
)

func TestIsMACConflict(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain", errors.New("boom"), false},
		{
			"mac-already-in-use",
			errors.New(`ServerError: 400 Bad Request ({"mac_addresses": ["MAC address aa:bb:cc:dd:ee:ff already in use on host-1."]})`),
			true,
		},
		{
			"all-already-exists",
			errors.New(`ServerError: 400 Bad Request ({"__all__": ["A node with this MAC address already exists."]})`),
			true,
		},
		{
			"list-already-in-use",
			errors.New(`ServerError: 400 Bad Request (["MAC address aa:bb:cc:dd:ee:ff already in use on host-1."])`),
			true,
		},
		{
			"400-but-unrelated",
			errors.New(`ServerError: 400 Bad Request ({"hostname": ["Invalid hostname."]})`),
			false,
		},
		{
			"500-error",
			errors.New(`ServerError: 500 Internal Server Error (something went wrong with mac already)`),
			false,
		},
		{
			"timeout",
			errors.New("dial tcp 1.2.3.4:80: i/o timeout"),
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isMACConflict(tc.err); got != tc.want {
				t.Fatalf("isMACConflict(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
