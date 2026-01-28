package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/poyrazk/terraform-provider-thecloud/internal/provider"
)

// Run the docs generation tool, check its usage at
// https://github.com/hashicorp/terraform-plugin-docs
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name thecloud

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/poyrazk/thecloud",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New("dev"), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
