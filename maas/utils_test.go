package maas

import (
	"fmt"
	"reflect"
	"testing"

	mock_gomaasapi "terraform-provider-maas/test/mocks"

	"github.com/golang/mock/gomock"
	"github.com/juju/gomaasapi"
	"github.com/stretchr/testify/assert"
)

func TestBase64Encode(t *testing.T) {
	testCases := []struct {
		name string
		in   []byte
		out  string
	}{
		// normal encoding case
		{
			name: "data is encoded",
			in:   []byte("data should be encoded"),
			out:  "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA==",
		},
		// base64 encoded input should result in no change of output
		{
			name: "data already encoded",
			in:   []byte("ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="),
			out:  "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA==",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			out := base64Encode(testCase.in)
			assert.Equal(t, testCase.out, out, fmt.Sprintf("base64Encode(%s) => %s, want %s", testCase.in, out, testCase.out))
		})
	}
}

func TestConvertToStringSlice(t *testing.T) {
	testCases := []struct {
		name string
		in   []interface{}
		out  []string
	}{
		{
			name: "empty slice",
			in:   []interface{}{},
			out:  []string{},
		},
		{
			name: "slice properly converted",
			in:   []interface{}{"elm1", "elem2", "elem3"},
			out:  []string{"elm1", "elem2", "elem3"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			out := convertToStringSlice(testCase.in)
			outType := reflect.TypeOf(out).Kind()
			assert.Equal(t, reflect.Slice, outType, fmt.Sprintf("convertToStringSlice(%s) has type %s, expected %s", testCase.in, outType, reflect.Slice))
			for i := range out {
				elemType := reflect.TypeOf(out[i]).Kind()
				assert.Equal(t, reflect.String, elemType, fmt.Sprintf("convertToStringSlice(%s)[%v] has type %s, expected %s", testCase.in, i, elemType, reflect.String))
			}
		})
	}
}

func TestGetMaasMachine(t *testing.T) {
	testCases := []struct {
		name        string
		machine_ids []string
		in          string
		err         error
	}{
		{
			name:        "machine is found",
			machine_ids: []string{"id-1"},
			in:          "id-1",
			err:         nil,
		},
		{
			name:        "machine is not found",
			machine_ids: []string{},
			in:          "id-1",
			err:         fmt.Errorf("machine (id-1) was not found"),
		},
		{
			name:        "multiple machines found",
			machine_ids: []string{"id-1", "id-2"},
			in:          "id-1",
			err:         fmt.Errorf("multiple machines found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			var machines []gomaasapi.Machine
			for _, id := range testCase.machine_ids {
				machine := mock_gomaasapi.NewMockMachine(mockCtrl)
				machine.
					EXPECT().
					SystemID().
					Return(id).
					AnyTimes()
				machines = append(machines, machine)
			}

			client := mock_gomaasapi.NewMockController(mockCtrl)
			client.
				EXPECT().
				Machines(gomock.Eq(gomaasapi.MachinesArgs{SystemIDs: []string{testCase.in}})).
				Return(machines, nil)

			ma, err := getMaasMachine(client, testCase.in)
			if testCase.err != nil {
				assert.Equal(t, testCase.err, err)
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, testCase.in, ma.SystemID())
				}
			}
		})
	}
}
