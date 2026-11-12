package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Requires the operator-created flag policy "num-shards-one"
// (description "Terraform Acceptance Test Flag Policy", flags "--num_shards=1").
func TestAcc_FlagPolicyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFlagPolicyDataSourceConfig("num-shards-one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dfcloud_flag_policy.test", "name", "num-shards-one"),
					resource.TestCheckResourceAttr("data.dfcloud_flag_policy.test", "description", "Terraform Acceptance Test Flag Policy"),
					resource.TestCheckResourceAttr("data.dfcloud_flag_policy.test", "flags.#", "1"),
					resource.TestCheckResourceAttr("data.dfcloud_flag_policy.test", "flags.0", "--num_shards=1"),
					resource.TestCheckResourceAttrSet("data.dfcloud_flag_policy.test", "id"),
					resource.TestCheckResourceAttrSet("data.dfcloud_flag_policy.test", "created_at"),
				),
			},
			{
				Config:      testAccFlagPolicyDataSourceConfig("does-not-exist-flag-policy"),
				ExpectError: regexp.MustCompile(`no flag policy named`),
			},
		},
	})
}

func testAccFlagPolicyDataSourceConfig(name string) string {
	return `
data "dfcloud_flag_policy" "test" {
  name = "` + name + `"
}
`
}
