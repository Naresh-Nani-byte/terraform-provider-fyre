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

# Example: Query VM status by VM ID
data "fyre_vm_status" "by_id" {
  vm_id = "1-8103661"
}

# Example: Query VM status by IP address
data "fyre_vm_status" "by_ip" {
  ip = "9.60.241.57"
}

# Example: Query VM status by FQDN
data "fyre_vm_status" "by_fqdn" {
  fqdn = "v1-8103661.dev.fyre.ibm.com"
}

output "vm_status_by_id" {
  description = "VM status queried by VM ID"
  value = {
    last_os_state = data.fyre_vm_status.by_id.last_os_state
    status        = data.fyre_vm_status.by_id.status
  }
}

output "vm_status_by_ip" {
  description = "VM status queried by IP address"
  value = {
    last_os_state = data.fyre_vm_status.by_ip.last_os_state
    status        = data.fyre_vm_status.by_ip.status
  }
}

output "vm_status_by_fqdn" {
  description = "VM status queried by FQDN"
  value = {
    last_os_state = data.fyre_vm_status.by_fqdn.last_os_state
    status        = data.fyre_vm_status.by_fqdn.status
  }
}
