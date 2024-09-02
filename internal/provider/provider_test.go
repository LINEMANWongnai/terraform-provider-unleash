package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/provider"
)

func init() {
	_ = os.Setenv("TF_ACC", "true")
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unleash": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
}

func getProviderConf(port int) string {
	return fmt.Sprintf(`
	provider "unleash" {
		  base_url = "http://localhost:%d"
		  authorization	= "*:development.x"
	}
	`, port)
}
