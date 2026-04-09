// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
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
				Config: testAccUserDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify top-level attributes
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "authenticated"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "email"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "login"),

					// Verify development nested object exists
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "development.id"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "development.username"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "development.email"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "development.display_name"),

					// Verify sentry nested object exists
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "sentry.status"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "sentry.access"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig() string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_user" "test" {}
`
}

func testAccPreCheck(t *testing.T) {
	// Check that required environment variables are set
	if v := os.Getenv("FYRE_USERNAME"); v == "" {
		t.Fatal("FYRE_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("FYRE_API_KEY"); v == "" {
		t.Fatal("FYRE_API_KEY must be set for acceptance tests")
	}
}
