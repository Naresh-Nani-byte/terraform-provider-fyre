# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    fyre = {
      source = "hashicorp-forge/fyre"
    }
  }
}

variable "product_group_id" {
  type        = number
  description = "The ID of the product group that you wish to create resources in"
}

provider "fyre" {
  # Username and API key can be set via FYRE_USERNAME and FYRE_API_KEY environment variables
  # or explicitly configured here:
  # username = "your-username"
  # api_key  = "your-api-key"


  # The default product_group_id to use for resources and datasources that
  # support it in their queries. When present at this level it will be inherited.
  # You can override it on a resource level by setting the product_group_id
  # attribute to a different value or null if you wish to not use it.
  # Can also be set with FYRE_PRODUCT_GROUP_ID
  product_group_id = var.product_group_id

  # Can also be set with FYRE_SITE
  site = "svl" # Default site (svl or rtp)
}

# Get available operating systems for x86 platform
data "fyre_vm_os_available" "x_platform" {
  platform = "x"
}

# Basic VM with minimal configuration
resource "fyre_vm" "basic" {
  os          = data.fyre_vm_os_available.x_platform.redhat[0]
  platform    = "x"
  cpu         = 2
  memory      = 4
  description = "Basic VM example"
}

# Check if a hostname is available
data "fyre_vm_check_hostname" "example" {
  hostname = "my-test-vm"
}

# VM with full configuration
resource "fyre_vm" "advanced" {
  os             = data.fyre_vm_os_available.x_platform.redhat[0]
  platform       = "x"
  cpu            = 2
  memory         = 4
  hostname       = data.fyre_vm_check_hostname.example.is_available ? data.fyre_vm_check_hostname.example.hostname : null
  description    = "Advanced VM with custom configuration"
  expiration     = "48" # 48 hours
  public_network = "y"
  dns            = "y"
  disable_delete = "n"
  quota_type     = "product_group" # or "quick_burn"

  # SSH key for access
  ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."

  # Or password
  # password    = "MySecurePassword123!"

  # Additional disks (max 2)
  additional_disks = ["50", "100"]
}

# Outputs
output "basic_vm_id" {
  description = "The VM ID of the basic VM"
  value       = fyre_vm.basic.vm_id
}

output "basic_vm_fqdn" {
  description = "The FQDN of the basic VM"
  value       = fyre_vm.basic.fqdn
}

output "advanced_vm_details" {
  description = "Full details of the advanced VM"
  value = {
    vm_id    = fyre_vm.advanced.vm_id
    fqdn     = fyre_vm.advanced.fqdn
    state    = fyre_vm.advanced.state
    location = fyre_vm.advanced.location
    created  = fyre_vm.advanced.created
  }
}
