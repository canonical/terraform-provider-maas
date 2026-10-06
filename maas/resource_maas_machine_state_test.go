package maas

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func TestMachineStateError(t *testing.T) {
	unexpected := &retry.UnexpectedStateError{
		State:         "Failed deployment",
		ExpectedState: []string{"Deployed"},
	}

	testCases := []struct {
		name    string
		waitErr error
		machine *entity.Machine
		want    []string
		absent  []string
	}{
		{
			name:    "an unexpected state carries MAAS's own explanation",
			waitErr: unexpected,
			machine: &entity.Machine{
				Hostname:      "acceptance-vm",
				SystemID:      "6fbce9ex",
				StatusMessage: "Failed to power on the machine",
			},
			want: []string{
				"acceptance-vm",
				"6fbce9ex",
				"Failed deployment",
				"Deployed",
				"Failed to power on the machine",
			},
			absent: []string{"%!s("},
		},
		{
			name:    "an empty status message says so",
			waitErr: unexpected,
			machine: &entity.Machine{Hostname: "acceptance-vm", SystemID: "6fbce9ex"},
			want: []string{
				"acceptance-vm",
				"Failed deployment",
				"MAAS reported no further detail",
			},
			absent: []string{"%!s("},
		},
		{
			name:    "timeouts are already described well",
			waitErr: &retry.TimeoutError{LastState: "Deploying", Timeout: time.Hour, ExpectedState: []string{"Deployed"}},
			machine: &entity.Machine{Hostname: "acceptance-vm"},
			want:    []string{"timeout while waiting", "Deploying"},
		},
		{
			name:    "plain errors are passed through",
			waitErr: errors.New("connection refused"),
			machine: &entity.Machine{Hostname: "acceptance-vm"},
			want:    []string{"connection refused"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := machineStateError(testCase.waitErr, testCase.machine, []string{"Deployed"})
			if err == nil {
				t.Fatal("machineStateError() = nil, want an error")
			}

			for _, want := range testCase.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("machineStateError() = %q, want it to contain %q", err, want)
				}
			}

			for _, absent := range testCase.absent {
				if strings.Contains(err.Error(), absent) {
					t.Errorf("machineStateError() = %q, want it not to contain %q", err, absent)
				}
			}
		})
	}
}
