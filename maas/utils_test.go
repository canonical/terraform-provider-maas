package maas

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBase64Encode(t *testing.T) {
	testCases := []struct {
		in  []byte
		out string
	}{
		// normal encoding case
		{[]byte("data should be encoded"), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
		// base64 encoded input should result in no change of output
		{[]byte("ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
	}

	for _, tt := range testCases {
		out := base64Encode(tt.in)
		if out != tt.out {
			t.Errorf("base64Encode(%s) => %s, want %s", tt.in, out, tt.out)
		}
	}
}

func TestConvertToStringSlice(t *testing.T) {
	testCases := []struct {
		in  []interface{}
		out []string
	}{
		{[]interface{}{}, []string{}},
		{[]interface{}{"elm1", "elem2", "elem3"}, []string{"elm1", "elem2", "elem3"}},
	}

	for _, tt := range testCases {
		out := convertToStringSlice(tt.in)
		outType := reflect.TypeOf(out).Kind()
		if outType != reflect.Slice {
			t.Errorf("convertToStringSlice(%s) has type %s, expected %s", tt.in, outType, reflect.Slice)
		}
		for i := range out {
			elemType := reflect.TypeOf(out[i]).Kind()
			if elemType != reflect.String {
				t.Errorf("convertToStringSlice(%s)[%v] has type %s, expected %s", tt.in, i, elemType, reflect.String)
			}
		}
	}
}

func TestGetMaasMachine(t *testing.T) {
	testCases := []struct {
		ctrl MockController
		in   string
		out  MockMachine
		err  error
	}{
		// normal case when machine is found, and returned
		{
			ctrl: MockController{
				machines: []MockMachine{
					{systemId: "id-1"},
				},
			},
			in:  "id-1",
			out: MockMachine{systemId: "id-1"},
			err: nil,
		},
		// machine is not found
		{
			ctrl: MockController{
				machines: []MockMachine{
					{systemId: "id-1"},
				},
			},
			in:  "id-4",
			out: MockMachine{},
			err: fmt.Errorf("machine (id-4) was not found"),
		},
		// multiple machines with same id are found
		{
			ctrl: MockController{
				machines: []MockMachine{
					{systemId: "id-1"},
					{systemId: "id-1"},
				},
			},
			in:  "id-1",
			out: MockMachine{},
			err: fmt.Errorf("multiple machines found"),
		},
	}

	for _, tt := range testCases {
		ma, err := getMaasMachine(&tt.ctrl, tt.in)
		if tt.err != nil {
			if tt.err.Error() != err.Error() {
				t.Errorf("getMaasMachine(&mockClient, %s) => err == '%s', expected err == '%s'", tt.in, err, tt.err)
			}
		} else if err != nil {
			t.Errorf("getMaasMachine(&mockClient, %s) => err == '%s', expected err == nil", tt.in, err)
		} else {
			if ma.SystemID() != tt.out.SystemID() {
				t.Errorf("getMaasMachine(&mockClient, %s) => machine (%s), expected machine (%s)", tt.in, ma.SystemID(), tt.out.SystemID())
			}
		}
	}
}
