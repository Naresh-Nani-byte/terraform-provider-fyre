// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceUserAPIKey(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccUserAPIKeyDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fyre_user_api_key.test", "id", "user_api_key"),
					resource.TestCheckResourceAttrSet("data.fyre_user_api_key.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_user_api_key.test", "status"),
					resource.TestCheckResourceAttrSet("data.fyre_user_api_key.test", "details"),
					resource.TestCheckResourceAttrSet("data.fyre_user_api_key.test", "expiration"),
				),
			},
		},
	})
}

func testAccUserAPIKeyDataSourceConfig() string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_user_api_key" "test" {
}
`
}
