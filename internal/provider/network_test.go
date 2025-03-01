package provider

import (
	"context"
	"fmt"
	"testing"

	dfcloud "github.com/dragonflydb/dfcloud/sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testCheckNetworkExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no network ID is set")
		}

		_, err := testClient().GetNetwork(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching network with ID %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testCheckNetworkDestroy(s *terraform.State) error {
	client := testClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dfcloud_network" {
			continue
		}

		network, err := client.GetNetwork(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching network with ID %s: %s", rs.Primary.ID, err)
		}

		if network.Status != dfcloud.NetworkStatusDeleting && network.Status != dfcloud.NetworkStatusDeleted {
			return fmt.Errorf("network still exists")
		}
	}

	return nil
}

func TestAcc_NetworkResource(t *testing.T) {
	name := "tf-test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckNetworkDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNetworkResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckNetworkExists("dfcloud_network.test"),
					resource.TestCheckResourceAttr("dfcloud_network.test", "name", name),
					resource.TestCheckResourceAttr("dfcloud_network.test", "location.provider", "aws"),
					resource.TestCheckResourceAttr("dfcloud_network.test", "location.region", "us-east-1"),
					resource.TestCheckResourceAttr("dfcloud_network.test", "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttrSet("dfcloud_network.test", "id"),
					resource.TestCheckResourceAttrSet("dfcloud_network.test", "created_at"),
					resource.TestCheckResourceAttrSet("dfcloud_network.test", "status"),
					resource.TestCheckResourceAttrSet("dfcloud_network.test", "vpc.resource_id"),
					resource.TestCheckResourceAttrSet("dfcloud_network.test", "vpc.account_id"),
				),
			},
			// Import State
			{
				ResourceName:      "dfcloud_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNetworkResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dfcloud_network" "test" {
  name = %[1]q
  
  location = {
	provider = "aws"
	region   = "us-east-1"
  }

  cidr_block = "10.0.0.0/16"
}
`, name)
}
