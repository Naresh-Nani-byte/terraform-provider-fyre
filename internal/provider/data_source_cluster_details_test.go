// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceClusterDetails tests the fyre_cluster_details data source
// with both include_vms=false and include_vms=true scenarios.
func TestAccDataSourceClusterDetails(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	clusterID := os.Getenv("FYRE_ACC_CLUSTER_ID")
	if clusterID == "" {
		t.Skip("FYRE_ACC_CLUSTER_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test without VMs
			{
				Config: testAccClusterDetailsDataSourceConfig(clusterID, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "id"),
					resource.TestCheckResourceAttr("data.fyre_cluster_details.test", "cluster_id", clusterID),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "site"),
					resource.TestCheckResourceAttr("data.fyre_cluster_details.test", "include_vms", "false"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "user_id"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "name"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "created"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "updated"),
				),
			},
			// Test with VMs
			{
				Config: testAccClusterDetailsDataSourceConfig(clusterID, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "id"),
					resource.TestCheckResourceAttr("data.fyre_cluster_details.test", "cluster_id", clusterID),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "site"),
					resource.TestCheckResourceAttr("data.fyre_cluster_details.test", "include_vms", "true"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "user_id"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "name"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "created"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "updated"),
					resource.TestCheckResourceAttrSet("data.fyre_cluster_details.test", "vms.#"),
				),
			},
		},
	})
}

func testAccClusterDetailsDataSourceConfig(clusterID string, includeVMs bool) string {
	if includeVMs {
		return `
provider "fyre" {
  site = "svl"
}

data "fyre_cluster_details" "test" {
  cluster_id  = "` + clusterID + `"
  include_vms = true
}
`
	}
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_cluster_details" "test" {
  cluster_id = "` + clusterID + `"
}
`
}
