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
  # Site can be set via FYRE_SITE environment variable or explicitly here
  site = "svl"
}

# Check if a hostname is available
data "fyre_vm_check_hostname" "example" {
  hostname = "my-test-vm"
}

# Output the availability status
output "hostname_available" {
  value       = data.fyre_vm_check_hostname.example.is_available
  description = "Whether the hostname is available for use"
}

output "hostname_details" {
  value       = data.fyre_vm_check_hostname.example.details
  description = "Details about the hostname availability"
}

# Conditional output - show FQDN if available
output "fqdn" {
  value       = data.fyre_vm_check_hostname.example.fqdn
  description = "Fully qualified domain name (only if hostname is available)"
}

# Conditional output - show owner if in use
output "owner_info" {
  value = data.fyre_vm_check_hostname.example.is_available ? null : {
    user_id  = data.fyre_vm_check_hostname.example.owning_user
    username = data.fyre_vm_check_hostname.example.owner.username
    email    = data.fyre_vm_check_hostname.example.owner.email
    vm_id    = data.fyre_vm_check_hostname.example.vm_id
  }
  description = "Owner information (only if hostname is in use)"
}
