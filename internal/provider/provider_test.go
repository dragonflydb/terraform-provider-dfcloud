package provider

import (
	"os"
	"testing"

	dfcloud "github.com/dragonflydb/dfcloud/sdk"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dfcloud": providerserver.NewProtocol6WithError(NewDragonflyDBCloudProvider("dev")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("DFCLOUD_API_KEY") == "" {
		t.Fatalf("DFCLOUD_API_KEY environment variable must be set for acceptance tests")
	}
}

var tc *dfcloud.Client

func testClient() *dfcloud.Client {
	if tc == nil {
		var options []dfcloud.ClientOption

		options = append(options, dfcloud.WithAPIKeyFromEnv())

		client, _ := dfcloud.NewClient(options...)

		tc = client
	}

	return tc
}
