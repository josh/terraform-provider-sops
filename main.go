package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version    string = "0.0.1"
	sopsBinary string = "sops"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/josh/sops",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
