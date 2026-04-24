// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceQuota verifies the fyre_quota data source can successfully
// retrieve quota information from the Fyre API. It requires FYRE_USERNAME and
// FYRE_API_KEY environment variables to be set. The test validates that the data
// source returns expected attributes including product group details, platform-specific
// resource quotas (CPU, memory, disk), and IP allocation quotas.
func TestAccDataSourceQuota(t *testing.T) {
	// Skip if credentials are not set
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDataSourceQuotaConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify top-level attributes
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "status"),

					// Verify details nested object exists
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.product_group_id"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.product_group_name"),

					// Verify X platform quotas
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.cpu"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.cpu_percent"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.cpu_used"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.disk"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.disk_percent"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.disk_used"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.memory"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.memory_percent"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.x.memory_used"),

					// Verify IP quotas
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.ip.public.quota"),
					resource.TestCheckResourceAttrSet("data.fyre_quota.test", "details.ip.public.used"),
				),
			},
		},
	})
}

func testAccDataSourceQuotaConfig() string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_quota" "test" {}
`
}
