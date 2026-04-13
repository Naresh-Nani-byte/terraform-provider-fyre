// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceVMStatus verifies the fyre_vm_status data source can successfully
// retrieve VM status information from the Fyre API. It requires FYRE_USERNAME,
// FYRE_API_KEY, and FYRE_TEST_VM_ID environment variables to be set.
// The test validates that the data source returns expected attributes including
// id, site, last_os_state, and status.
func TestAccDataSourceVMStatus(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	vmID := os.Getenv("FYRE_TEST_VM_ID")
	if vmID == "" {
		t.Skip("FYRE_TEST_VM_ID must be set for VM status acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with VM ID
			{
				Config: testAccVMStatusDataSourceConfig(vmID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_vm_status.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_status.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_status.test", "last_os_state"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_status.test", "status"),
				),
			},
		},
	})
}

func testAccVMStatusDataSourceConfig(vmID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

data "fyre_vm_status" "test" {
  vm_id = %[1]q
}
`, vmID)
}
