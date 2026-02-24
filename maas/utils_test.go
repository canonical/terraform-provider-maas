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

func TestCheckSemverConstraint(t *testing.T) {
	tests := []struct {
		name             string
		MAASVersion      string
		semverConstraint string
		expectedErr      bool
		ErrString        string
	}{
		{
			name:             ">= constraint",
			MAASVersion:      "2.0.0",
			semverConstraint: ">= 2.0.0",
			expectedErr:      false,
		},
		{
			name:             "== constraint",
			MAASVersion:      "3.1.4",
			semverConstraint: "= 3.1.4",
			expectedErr:      false,
		},
		{
			name:             "<= constraint",
			MAASVersion:      "3.1.3",
			semverConstraint: "<= 3.1.4",
			expectedErr:      false,
		},
		{
			name:             "invalid constraint",
			MAASVersion:      "1.0.0",
			semverConstraint: ">= 2.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `1.0.0`, does not satisfy constraint `>= 2.0.0`",
		},
		{
			name:             "empty constraint",
			MAASVersion:      "1.0.0",
			semverConstraint: "",
			expectedErr:      false,
		},
		{
			name:             "empty version",
			MAASVersion:      "",
			semverConstraint: ">= 2.0.0",
			expectedErr:      false,
		},
		{
			name:             "invalid constraint format",
			MAASVersion:      "3.0.0",
			semverConstraint: ">> 2.0.0",
			expectedErr:      true,
			ErrString:        "improper constraint: >> 2.0.0",
		},
		{
			name:             "debian-style version with tilde below constraint",
			MAASVersion:      "2.9.0~alpha",
			semverConstraint: ">= 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `2.9.0~alpha`, does not satisfy constraint `>= 3.0.0`",
		},
		{
			name:             "debian-style version with tilde satisfies stable constraint",
			MAASVersion:      "3.1.0~beta1",
			semverConstraint: ">= 3.0.0",
			expectedErr:      false,
		},
		{
			name:             "debian-style version satisfies range constraint",
			MAASVersion:      "3.5.0~rc1",
			semverConstraint: ">= 3.0.0, < 4.0.0",
			expectedErr:      false,
		},
		{
			name:             "debian-style version fails constraint",
			MAASVersion:      "2.5.0~beta",
			semverConstraint: ">= 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `2.5.0~beta`, does not satisfy constraint `>= 3.0.0`",
		},
		{
			name:             "debian-style prerelease at boundary does not satisfy constraint",
			MAASVersion:      "3.0.0~rc1",
			semverConstraint: ">= 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `3.0.0~rc1`, does not satisfy constraint `>= 3.0.0`",
		},
		{
			name:             "4.0.0~rc1 is less than 4.0.0",
			MAASVersion:      "4.0.0~rc1",
			semverConstraint: "< 4.0.0",
			expectedErr:      false,
		},
		{
			name:             "3.1.0~beta is greater than 3.0.0",
			MAASVersion:      "3.1.0~beta",
			semverConstraint: "> 3.0.0",
			expectedErr:      false,
		},
		{
			name:             "3.0.0~rc1 satisfies <= 3.0.0",
			MAASVersion:      "3.0.0~rc1",
			semverConstraint: "<= 3.0.0",
			expectedErr:      false,
		},
		{
			name:             "3.1.0~beta fails <= 3.0.0",
			MAASVersion:      "3.1.0~beta",
			semverConstraint: "<= 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `3.1.0~beta`, does not satisfy constraint `<= 3.0.0`",
		},
		{
			name:             "stable > passes",
			MAASVersion:      "3.1.0",
			semverConstraint: "> 3.0.0",
			expectedErr:      false,
		},
		{
			name:             "stable > fails",
			MAASVersion:      "2.5.0",
			semverConstraint: "> 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `2.5.0`, does not satisfy constraint `> 3.0.0`",
		},
		{
			name:             "stable < passes",
			MAASVersion:      "2.5.0",
			semverConstraint: "< 3.0.0",
			expectedErr:      false,
		},
		{
			name:             "stable < fails",
			MAASVersion:      "3.5.0",
			semverConstraint: "< 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `3.5.0`, does not satisfy constraint `< 3.0.0`",
		},
		{
			name:             "pre-release > at boundary fails",
			MAASVersion:      "3.0.0~rc1",
			semverConstraint: "> 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `3.0.0~rc1`, does not satisfy constraint `> 3.0.0`",
		},
		{
			name:             "pre-release < below boundary passes",
			MAASVersion:      "2.5.0~beta",
			semverConstraint: "< 3.0.0",
			expectedErr:      false,
		},
		{
			name:             "pre-release < above boundary fails",
			MAASVersion:      "4.0.0~beta",
			semverConstraint: "< 3.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `4.0.0~beta`, does not satisfy constraint `< 3.0.0`",
		},
		{
			name:             "stable in range passes",
			MAASVersion:      "3.5.0",
			semverConstraint: ">= 3.0.0, < 4.0.0",
			expectedErr:      false,
		},
		{
			name:             "stable outside range fails",
			MAASVersion:      "4.5.0",
			semverConstraint: ">= 3.0.0, < 4.0.0",
			expectedErr:      true,
			ErrString:        "MAAS version `4.5.0`, does not satisfy constraint `>= 3.0.0, < 4.0.0`",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := checkSemverConstraint(testCase.MAASVersion, testCase.semverConstraint)
			if (err != nil) != testCase.expectedErr {
				t.Errorf("checkSemverConstraint() error = %v, expectedErr %v", err, testCase.expectedErr)
			}

			if err != nil && err.Error() != testCase.ErrString {
				t.Errorf("checkSemverConstraint() error = %v, ErrString %v", err.Error(), testCase.ErrString)
			}
		})
	}
}

func TestListAsString(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{
			name:     "empty list",
			input:    []any{},
			expected: "",
		},
		{
			name:     "single element",
			input:    []any{"foo"},
			expected: "foo",
		},
		{
			name:     "multiple elements",
			input:    []any{"foo", "bar", "baz"},
			expected: "foo,bar,baz",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := listAsString(testCase.input)
			if result != testCase.expected {
				t.Errorf("listAsString() result = %v, expected %v", result, testCase.expected)
			}
		})
	}
}

func TestOptionalStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-empty string",
			input:    "foo",
			expected: func() *string { s := "foo"; return &s }(),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := optionalStringPtr(testCase.input)
			switch {
			case result == nil && testCase.expected != nil:
				t.Errorf("optionalStringPtr() result = nil, expected %v", *testCase.expected)
			case result != nil && testCase.expected == nil:
				t.Errorf("optionalStringPtr() result = %v, expected nil", *result)
			case result != nil && testCase.expected != nil && *result != *testCase.expected:
				t.Errorf("optionalStringPtr() result = %v, expected %v", *result, *testCase.expected)
			}
		})
	}
}
