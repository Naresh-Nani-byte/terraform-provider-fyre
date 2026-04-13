// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVMSnapshotsDataSource(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	vmID := os.Getenv("FYRE_TEST_VM_ID")
	if vmID == "" {
		t.Skip("FYRE_TEST_VM_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMSnapshotsDataSourceConfig(vmID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_vm_snapshots.test", "id"),
					resource.TestCheckResourceAttr("data.fyre_vm_snapshots.test", "vm_id", vmID),
					resource.TestCheckResourceAttrSet("data.fyre_vm_snapshots.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_snapshots.test", "snapshot_count"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_snapshots.test", "snapshot_limit"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_snapshots.test", "snapshots.#"),
				),
			},
		},
	})
}

func testAccVMSnapshotsDataSourceConfig(vmID string) string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_vm_snapshots" "test" {
  vm_id = "` + vmID + `"
}
`
}
