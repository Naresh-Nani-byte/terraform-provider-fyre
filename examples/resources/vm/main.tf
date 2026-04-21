# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    fyre = {
      source = "hashicorp-forge/fyre"
    }
  }
}

provider "fyre" {
  # Username and API key can be set via FYRE_USERNAME and FYRE_API_KEY environment variables
  # or explicitly configured here:
  # username = "your-username"
  # api_key  = "your-api-key"
  
  site = "svl"  # Default site (svl or rtp)
}

# Basic VM with minimal configuration
resource "fyre_vm" "basic" {
  os          = "RHEL 9.6"
  cpu         = 2
  memory      = 4
  description = "Basic VM example"
}

# VM with full configuration
resource "fyre_vm" "advanced" {
  os               = "RHEL 9.6"
  platform         = "x"
  cpu              = 4
  memory           = 8
  hostname         = "myvm"
  description      = "Advanced VM with custom configuration"
  expiration       = "48"  # 48 hours
  public_network   = "y"
  dns              = "y"
  disable_delete   = "n"
  quota_type       = "product_group"
  product_group_id = "your-product-group-id"
  
  # SSH key for access
  ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
  
  # Additional disks (max 2)
  additional_disks = ["50", "100"]
}

# VM with password authentication
resource "fyre_vm" "with_password" {
  os          = "Ubuntu 22.04"
  cpu         = 2
  memory      = 4
  description = "VM with custom password"
  password    = "MySecurePassword123!"
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
