package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-maas/maas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: maas.Provider}

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/canonical/maas", opts)

		if err != nil {
			log.Fatal(err.Error())
		}

		return
	}

	plugin.Serve(opts)
}
