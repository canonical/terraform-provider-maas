package main

import (
	"flag"
	"terraform-provider-maas/maas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: maas.Provider}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/canonical/maas"
	}

	plugin.Serve(opts)
}
