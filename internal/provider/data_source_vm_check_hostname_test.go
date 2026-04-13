// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceVMCheckHostname verifies the fyre_vm_check_hostname data source can
// successfully check hostname availability in the Fyre environment. It requires
// FYRE_USERNAME, FYRE_API_KEY, and FYRE_ACC_VM_ID environment variables to be set.
// The test validates both available and in-use hostname scenarios, checking that
// appropriate attributes are returned for each case.
func TestAccDataSourceVMCheckHostname(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	vmID := os.Getenv("FYRE_ACC_VM_ID")
	if vmID == "" {
		t.Skip("FYRE_ACC_VM_ID must be set for acceptance tests")
	}

	// Get the hostname from the VM details using the API
	username := os.Getenv("FYRE_USERNAME")
	apiKey := os.Getenv("FYRE_API_KEY")

	basicAuthEditor := func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth(username, apiKey)
		return nil
	}

	fyreClient, err := client.NewClientWithResponses(
		client.ServerUrlSVLSanJoseProductionServer,
		client.WithRequestEditorFn(basicAuthEditor),
	)
	if err != nil {
		t.Fatalf("Failed to create Fyre client: %v", err)
	}

	site := client.GetVMDetailsParamsSite("svl")
	vmResp, err := fyreClient.GetVMDetailsWithResponse(context.Background(), client.VmIdentifier(vmID), &client.GetVMDetailsParams{
		Site: &site,
	})
	if err != nil {
		t.Fatalf("Failed to get VM details: %v", err)
	}

	if vmResp.StatusCode() != 200 || vmResp.JSON200 == nil {
		t.Fatalf("Failed to get VM details: status %d", vmResp.StatusCode())
	}

	if vmResp.JSON200.Hostname == nil {
		t.Fatal("VM hostname is nil")
	}

	inUseHostname := *vmResp.JSON200.Hostname

	// Use a known available hostname for the available test
	availableHostname := "not-in-use"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with available hostname
			{
				Config: testAccDataSourceVMCheckHostnameConfigAvailable(availableHostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.available", "id"),
					resource.TestCheckResourceAttr("data.fyre_vm_check_hostname.available", "hostname", availableHostname),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.available", "site"),
					resource.TestCheckResourceAttr("data.fyre_vm_check_hostname.available", "status", "success"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.available", "details"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.available", "fqdn"),
					resource.TestCheckResourceAttr("data.fyre_vm_check_hostname.available", "is_available", "true"),
				),
			},
			// Test with in-use hostname (from actual VM)
			{
				Config: testAccDataSourceVMCheckHostnameConfigInUse(inUseHostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "id"),
					resource.TestCheckResourceAttr("data.fyre_vm_check_hostname.in_use", "hostname", inUseHostname),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "site"),
					resource.TestCheckResourceAttr("data.fyre_vm_check_hostname.in_use", "status", "warning"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "details"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "owning_user"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "owner.id"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "owner.username"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "owner.email"),
					resource.TestCheckResourceAttrSet("data.fyre_vm_check_hostname.in_use", "vm_id"),
					resource.TestCheckResourceAttr("data.fyre_vm_check_hostname.in_use", "is_available", "false"),
				),
			},
		},
	})
}

func testAccDataSourceVMCheckHostnameConfigAvailable(hostname string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

data "fyre_vm_check_hostname" "available" {
  hostname = %[1]q
}
`, hostname)
}

func testAccDataSourceVMCheckHostnameConfigInUse(hostname string) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

data "fyre_vm_check_hostname" "in_use" {
  hostname = %[1]q
}
`, hostname)
}