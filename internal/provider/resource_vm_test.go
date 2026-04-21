// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceVM runs the acceptance test for creating, updating, and
// destroying VM's in Fyre.
func TestAccResourceVM(t *testing.T) {
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
			// Create and Read testing
			{
				Config: testAccResourceVMConfigBasic("RedHat 9.6", 2, 4, "Test VM for Terraform acceptance testing", "24", "n", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "os", "RedHat 9.6"),
					resource.TestCheckResourceAttr("fyre_vm.test", "cpu", "2"),
					resource.TestCheckResourceAttr("fyre_vm.test", "memory", "4"),
					resource.TestCheckResourceAttr("fyre_vm.test", "description", "Test VM for Terraform acceptance testing"),
					resource.TestCheckResourceAttr("fyre_vm.test", "expiration", "24"),
					resource.TestCheckResourceAttr("fyre_vm.test", "disable_delete", "n"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "id"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "vm_id"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "site"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "platform"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "expiration_time"),
				),
			},
			// Update CPU and Memory
			{
				Config: testAccResourceVMConfigBasic("RedHat 9.6", 4, 8, "Test VM for Terraform acceptance testing", "24", "n", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "cpu", "4"),
					resource.TestCheckResourceAttr("fyre_vm.test", "memory", "8"),
					resource.TestCheckResourceAttr("fyre_vm.test", "description", "Test VM for Terraform acceptance testing"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "id"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "vm_id"),
				),
			},
			// Update Description
			{
				Config: testAccResourceVMConfigBasic("RedHat 9.6", 4, 8, "Updated test VM description", "24", "n", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "cpu", "4"),
					resource.TestCheckResourceAttr("fyre_vm.test", "memory", "8"),
					resource.TestCheckResourceAttr("fyre_vm.test", "description", "Updated test VM description"),
				),
			},
			// Update Expiration
			{
				Config: testAccResourceVMConfigBasic("RedHat 9.6", 4, 8, "Updated test VM description", "48", "n", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "expiration", "48"),
				),
			},
			// Update Disable Delete
			{
				Config: testAccResourceVMConfigBasic("RedHat 9.6", 4, 8, "Updated test VM description", "48", "y", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "disable_delete", "y"),
				),
			},
			// Update Password
			{
				Config: testAccResourceVMConfigWithPassword("RedHat 9.6", 4, 8, "Updated test VM description", "48", "y", "NewTestPassword123!", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "password", "NewTestPassword123!"),
				),
			},
			// Update Additional Disks
			{
				Config: testAccResourceVMConfigWithDisks("RedHat 9.6", 4, 8, "Updated test VM description", "48", "y", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "additional_disks.#", "2"),
					resource.TestCheckResourceAttr("fyre_vm.test", "additional_disks.0", "50"),
					resource.TestCheckResourceAttr("fyre_vm.test", "additional_disks.1", "100"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "fyre_vm.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore fields that are not returned by the API or are sensitive
				// expiration: user input (relative time) not returned by API, only expiration_time (absolute) is returned
				ImportStateVerifyIgnore: []string{"password", "ssh_key", "expiration"},
			},
			// Re-enable delete so the test sweeper can actually clean up the resource
			{
				Config: testAccResourceVMConfigWithDisks("RedHat 9.6", 4, 8, "Updated test VM description", "48", "n", productGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "disable_delete", "n"),
				),
			},
		},
	})
}

func testAccResourceVMConfigBasic(os string, cpu, memory int, description, expiration, disableDelete, productGroupID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

resource "fyre_vm" "test" {
  os               = %[1]q
  cpu              = %[2]d
  memory           = %[3]d
  description      = %[4]q
  platform         = "x"
  expiration       = %[5]q
  disable_delete   = %[6]q
  product_group_id = %[7]q
}
`, os, cpu, memory, description, expiration, disableDelete, productGroupID)
}

func testAccResourceVMConfigWithPassword(os string, cpu, memory int, description, expiration, disableDelete, password, productGroupID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

resource "fyre_vm" "test" {
  os               = %[1]q
  cpu              = %[2]d
  memory           = %[3]d
  description      = %[4]q
  platform         = "x"
  expiration       = %[5]q
  disable_delete   = %[6]q
  password         = %[7]q
  product_group_id = %[8]q
}
`, os, cpu, memory, description, expiration, disableDelete, password, productGroupID)
}

func testAccResourceVMConfigWithDisks(os string, cpu, memory int, description, expiration, disableDelete, productGroupID string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

resource "fyre_vm" "test" {
  os               = %[1]q
  cpu              = %[2]d
  memory           = %[3]d
  description      = %[4]q
  platform         = "x"
  expiration       = %[5]q
  disable_delete   = %[6]q
  additional_disks = ["50", "100"]
  product_group_id = %[7]q
}
`, os, cpu, memory, description, expiration, disableDelete, productGroupID)
}
