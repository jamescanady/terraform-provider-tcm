package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/symplr-software/terraform-provider-tcm/internal"
)

var version string = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run with delve debugger support")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/symplr/tcm",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), internal.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
