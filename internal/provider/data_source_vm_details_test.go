// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceVMDetails verifies the fyre_vm_details data source can successfully
// retrieve comprehensive VM information from the Fyre API. It requires FYRE_USERNAME,
// FYRE_API_KEY, and FYRE_ACC_VM_ID environment variables to be set. The test validates
// that the data source returns expected attributes including VM configuration, resource
// allocation, networking details, owner information, and operational status.
func TestAccDataSourceVMDetails(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	vmID := os.Getenv("FYRE_ACC_VM_ID")
	if vmID == "" {
		t.Skip("FYRE_ACC_VM_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with VM ID
			{
				Config: testAccDataSourceVMDetailsConfigByVMID(vmID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "vm_id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "location"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "hostname"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "fqdn"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "state"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "platform"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "os"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "cpu"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "memory"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "os_disk"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "quota_type"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "product_group_id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "product_group"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "created"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "user.id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "user.username"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "user.email"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_details.test", "ips.#"),
				),
			},
		},
	})
}

func testAccDataSourceVMDetailsConfigByVMID(vmID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

data "fyre_vm_details" "test" {
  vm_id = %q
}
`, vmID)
}
