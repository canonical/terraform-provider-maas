package maas_test

import (
	"terraform-provider-maas/maas"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := maas.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = maas.Provider()
}
