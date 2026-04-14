// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceClusters tests the fyre_clusters data source.
// This test requires FYRE_USERNAME and FYRE_API_KEY environment variables to be set.
// The test validates that the data source can successfully fetch the list of clusters
// for the authenticated user.
func TestAccDataSourceClusters(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_clusters.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_clusters.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_clusters.test", "cluster_count"),
					resource.TestCheckResourceAttrSet("data.fyre_clusters.test", "clusters.#"),
				),
			},
		},
	})
}

func testAccClustersDataSourceConfig() string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_clusters" "test" {
}
`
}
