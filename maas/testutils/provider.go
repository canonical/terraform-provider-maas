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
