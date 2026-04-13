// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVMOSAvailable(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test x platform
			{
				Config: testAccDataSourceVMOSAvailableConfig("x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fyre_vm_os_available.test", "platform", "x"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "operating_systems.%"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.count"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.cpu"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.memory"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.max_count"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.max_cpu"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.max_memory"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.max_disk_count"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.max_disk_size"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.max_total_disk_size"),
				),
			},
			// Test z platform
			{
				Config: testAccDataSourceVMOSAvailableConfig("z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fyre_vm_os_available.test", "platform", "z"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "operating_systems.%"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "default_size.count"),
				),
			},
			// Test with explicit site parameter
			{
				Config: testAccDataSourceVMOSAvailableConfigWithSite("x", "rtp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fyre_vm_os_available.test", "platform", "x"),
					resource.TestCheckResourceAttr("data.fyre_vm_os_available.test", "site", "rtp"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_os_available.test", "operating_systems.%"),
				),
			},
		},
	})
}

func testAccDataSourceVMOSAvailableConfig(platform string) string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_vm_os_available" "test" {
  platform = "` + platform + `"
}
`
}

func testAccDataSourceVMOSAvailableConfigWithSite(platform, site string) string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_vm_os_available" "test" {
  platform = "` + platform + `"
  site     = "` + site + `"
}
`
}
