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
  site = "svl"
}

# Example: Get snapshots for a VM using VM ID
data "fyre_vm_snapshots" "example" {
  vm_id = "1-8103661"
}

# Example: Get snapshots using IP address
data "fyre_vm_snapshots" "by_ip" {
  vm_id = "10.16.23.45"
}

# Example: Get snapshots using FQDN
data "fyre_vm_snapshots" "by_fqdn" {
  vm_id = "myvm.dev.fyre.ibm.com"
}

# Output the snapshot information
output "snapshot_count" {
  value       = data.fyre_vm_snapshots.example.snapshot_count
  description = "Number of snapshots for the VM"
}

output "snapshot_limit" {
  value       = data.fyre_vm_snapshots.example.snapshot_limit
  description = "Maximum number of snapshots allowed"
}

output "snapshots" {
  value       = data.fyre_vm_snapshots.example.snapshots
  description = "List of all snapshots with details"
}

# Example: Access individual snapshot details
output "first_snapshot_name" {
  value       = length(data.fyre_vm_snapshots.example.snapshots) > 0 ? data.fyre_vm_snapshots.example.snapshots[0].name : null
  description = "Name of the first snapshot if it exists"
}
