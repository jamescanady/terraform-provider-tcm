package internal_test

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/symplr-software/terraform-provider-tcm/internal"
)

func providerFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"tcm": providerserver.NewProtocol6WithError(internal.New("test")()),
	}
}
