package main

import (
	"context"
	"flag"
	"log"

	"github.com/dragonflydb/terraform-provider-dfcloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// version is the version of the provider.
	// This is set at compile time using -ldflags.
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/dragonflydb/dfcloud",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.NewDragonflyDBCloudProvider(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
