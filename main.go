package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"gitlab.com/ptrck-sh/terraform-provider-plane/internal/provider"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name plane

// version is set at release time via -ldflags by GoReleaser.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/ptrck-sh/plane",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
