package maas

import (
	"fmt"
	"reflect"
	"testing"

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
		in   []any
		out  []string
	}{
		{
			name: "empty slice",
			in:   []any{},
			out:  []string{},
		},
		{
			name: "slice properly converted",
			in:   []any{"elm1", "elem2", "elem3"},
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

func TestSplitStateIDIntoInts(t *testing.T) {
	tests := []struct {
		name        string
		stateID     string
		delimeter   string
		expectedID1 int
		expectedID2 int
		expectedErr bool
	}{
		{
			name:        "valid state ID with forward slash",
			stateID:     "123/456",
			delimeter:   "/",
			expectedID1: 123,
			expectedID2: 456,
			expectedErr: false,
		},
		{
			name:        "valid state ID with colon",
			stateID:     "123:456",
			delimeter:   ":",
			expectedID1: 123,
			expectedID2: 456,
			expectedErr: false,
		},
		{
			name:        "invalid state ID format",
			stateID:     "123",
			delimeter:   "/",
			expectedErr: true,
		},
		{
			name:        "non-integer values",
			stateID:     "abc/def",
			delimeter:   "/",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resID1, resID2, err := SplitStateIDIntoInts(tt.stateID, tt.delimeter)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("SplitStateIDIntoInts() error = nil, expectedErr %v", tt.expectedErr)
				}

				return
			}

			if err != nil {
				t.Errorf("SplitStateIDIntoInts() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if resID1 != tt.expectedID1 {
				t.Errorf("SplitStateIDIntoInts() resID1 = %v, expected %v", resID1, tt.expectedID1)
			}

			if resID2 != tt.expectedID2 {
				t.Errorf("SplitStateIDIntoInts() resID2 = %v, expected %v", resID2, tt.expectedID2)
			}
		})
	}
}
