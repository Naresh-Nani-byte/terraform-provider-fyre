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

# Example: Get cluster details without VMs
data "fyre_cluster_details" "example" {
  cluster_id = "15281"
}

output "cluster_info" {
  value = data.fyre_cluster_details.example
}

# Example: Get cluster details with VMs included
data "fyre_cluster_details" "with_vms" {
  cluster_id  = "15281"
  include_vms = true
}

output "cluster_with_vms" {
  value = data.fyre_cluster_details.with_vms
}
