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
