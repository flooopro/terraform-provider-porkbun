// main.go
package main

import (
	"context"
	"log"

	"github.com/flooopro/terraform-provider-porkbun/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version string = "0.0.1"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "github.com/flooopro/porkbun",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
