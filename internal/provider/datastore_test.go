package provider

import (
	"context"
	"fmt"
	"testing"

	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testCheckDatastoreExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no datastore ID is set")
		}

		_, err := testClient().GetDatastore(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching datastore with ID %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testCheckDatastoreDestroy(s *terraform.State) error {
	client := testClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dfcloud_datastore" {
			continue
		}

		ds, err := client.GetDatastore(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching datastore with ID %s: %s", rs.Primary.ID, err)
		}

		if ds.Status != dfcloud.DatastoreStatusDeleting && ds.Status != dfcloud.DatastoreStatusDeleted {
			return fmt.Errorf("datastore still exists")
		}
	}

	return nil
}

func TestAcc_DatastoreResource(t *testing.T) {
	name := "tf-test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatastoreDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatastoreResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckDatastoreExists("dfcloud_datastore.test"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "name", name),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "location.provider", "aws"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "location.region", "eu-west-1"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "location.availability_zones.#", "1"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "location.availability_zones.0", "euw1-az2"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "tier.performance_tier", "dev"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "tier.max_memory_bytes", "3000000000"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "tier.replicas", "1"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "dragonfly.cache_mode", "false"),
					resource.TestCheckResourceAttr("dfcloud_datastore.test", "dragonfly.tls", "false"),
					resource.TestCheckResourceAttrSet("dfcloud_datastore.test", "id"),
					resource.TestCheckResourceAttrSet("dfcloud_datastore.test", "addr"),
					resource.TestCheckResourceAttrSet("dfcloud_datastore.test", "created_at"),
					resource.TestCheckResourceAttrSet("dfcloud_datastore.test", "password"),
				),
			},
			// Import State
			{
				ResourceName:      "dfcloud_datastore.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDatastoreResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dfcloud_datastore" "test" {
  name = %[1]q
  
  location = {
    provider = "aws"
    region   = "eu-west-1"
	availability_zones = ["euw1-az2"]
  }

  tier = {
    max_memory_bytes  = 3000000000  # 3GB
    performance_tier = "dev"
    replicas        = 1
  }
}
`, name)
}
