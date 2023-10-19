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

func PreCheck(t *testing.T) {
	if v := os.Getenv("MAAS_API_URL"); v == "" {
		t.Fatal("MAAS_API_URL must be set for acceptance tests")
	}
	if v := os.Getenv("MAAS_API_KEY"); v == "" {
		t.Fatal("MAAS_API_KEY must be set for acceptance tests")
	}
}
