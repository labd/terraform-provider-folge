package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/labd/terraform-provider-folge/internal"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name folge

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/labd/folge",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), func() provider.Provider {
		var options = []internal.OptionFunc{
			//We allow 10 retries of a failed request
			internal.WithRetryableClient(10),
			internal.WithDebugClient(),
		}

		return internal.New(options...)
	}, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
