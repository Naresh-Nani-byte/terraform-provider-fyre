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

# Example: Query VM details by VM ID
data "fyre_vm_details" "by_id" {
  vm_id = "1-8103661"
}

# Example: Query VM details by IP address
data "fyre_vm_details" "by_ip" {
  ip = "9.60.241.57"
}

# Example: Query VM details by FQDN
data "fyre_vm_details" "by_fqdn" {
  fqdn = "v1-8103661.dev.fyre.ibm.com"
}

output "vm_details_by_id" {
  description = "Complete VM details queried by VM ID"
  value       = data.fyre_vm_details.by_id
}

output "vm_basic_info" {
  description = "Basic VM information"
  value = {
    hostname = data.fyre_vm_details.by_id.hostname
    fqdn     = data.fyre_vm_details.by_id.fqdn
    state    = data.fyre_vm_details.by_id.state
    platform = data.fyre_vm_details.by_id.platform
    os       = data.fyre_vm_details.by_id.os
  }
}

output "vm_resources" {
  description = "VM resource allocation"
  value = {
    cpu     = data.fyre_vm_details.by_id.cpu
    memory  = data.fyre_vm_details.by_id.memory
    os_disk = data.fyre_vm_details.by_id.os_disk
  }
}

output "vm_owner" {
  description = "VM owner information"
  value       = data.fyre_vm_details.by_id.user
}

output "vm_ips" {
  description = "VM IP addresses"
  value       = data.fyre_vm_details.by_id.ips
}
