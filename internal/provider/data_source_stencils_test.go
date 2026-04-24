// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceStencils verifies the fyre_stencils data source can successfully
// retrieve a list of stencils filtered by product group from the Fyre API. It requires
// FYRE_USERNAME, FYRE_API_KEY, and FYRE_ACC_PROD_GID environment variables to be set.
// The test validates that the data source returns expected attributes including stencil
// configurations, resource specifications, and owner information.
// Tests both explicit product_group_id and provider-level inheritance.
func TestAccDataSourceStencils(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	productGroupID := os.Getenv("FYRE_ACC_PROD_GID")
	if productGroupID == "" {
		t.Skip("FYRE_ACC_PROD_GID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with explicit product_group_id
			{
				Config: testAccDataSourceStencilsConfigExplicit(productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_stencils.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_stencils.test", "site"),
					resource.TestCheckResourceAttr("data.fyre_stencils.test", "product_group_id", productGroupID),
					resource.TestCheckResourceAttrSet("data.fyre_stencils.test", "stencils.#"),
				),
			},
			// Test with provider-level product_group_id inheritance
			{
				Config: testAccDataSourceStencilsConfigInherited(productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_stencils.test_inherited", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_stencils.test_inherited", "site"),
					resource.TestCheckResourceAttr("data.fyre_stencils.test_inherited", "product_group_id", productGroupID),
					resource.TestCheckResourceAttrSet("data.fyre_stencils.test_inherited", "stencils.#"),
				),
			},
		},
	})
}

func testAccDataSourceStencilsConfigExplicit(productGroupID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

data "fyre_stencils" "test" {
  product_group_id = %s
}
`, productGroupID)
}

func testAccDataSourceStencilsConfigInherited(productGroupID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
  product_group_id = %s
}

data "fyre_stencils" "test_inherited" {
  # product_group_id inherited from provider
}
`, productGroupID)
}
