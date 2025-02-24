package testutils

import (
	"os"
	"terraform-provider-maas/maas"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	TestAccProviders map[string]*schema.Provider
	TestAccProvider  *schema.Provider
)

func init() {
	TestAccProvider = maas.Provider()
	TestAccProviders = map[string]*schema.Provider{
		"maas": TestAccProvider,
	}
}

// PreCheck verifies that required environment variables are set before running acceptance tests.
// It takes a testing.T instance and an optional slice of additional required variables.
// The presence of MAAS_API_URL and MAAS_API_KEY are checked by default.
// If any required variables are missing, the test fails.
func PreCheck(t *testing.T, extra []string) {
	var requiredVariables = []string{"MAAS_API_URL", "MAAS_API_KEY"}
	missingVariables := new([]string)

	for _, rv := range append(requiredVariables, extra...) {
		if v := os.Getenv(rv); v == "" {
			*missingVariables = append(*missingVariables, rv)
		}
	}

	if len(*missingVariables) > 0 {
		t.Fatalf("%s must be set for acceptance tests", *missingVariables)
	}
}
