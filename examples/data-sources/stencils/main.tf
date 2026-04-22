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

# Product group ID can be set at provider level via FYRE_PRODUCT_GROUP_ID environment variable
# or passed directly to the data source
data "fyre_stencils" "example" {
  product_group_id = 1
}

output "stencils" {
  value = data.fyre_stencils.example.stencils
}

output "stencil_count" {
  value = length(data.fyre_stencils.example.stencils)
}

output "stencil_names" {
  description = "List of stencil names"
  value       = [for s in data.fyre_stencils.example.stencils : s.name]
}

output "stencil_platforms" {
  description = "List of platforms available"
  value       = distinct([for s in data.fyre_stencils.example.stencils : s.platform])
}

output "first_stencil_details" {
  description = "Detailed information about the first stencil"
  value = length(data.fyre_stencils.example.stencils) > 0 ? {
    name        = data.fyre_stencils.example.stencils[0].name
    description = data.fyre_stencils.example.stencils[0].description
    platform    = data.fyre_stencils.example.stencils[0].platform
    os          = data.fyre_stencils.example.stencils[0].os
    cpu         = data.fyre_stencils.example.stencils[0].cpu
    memory      = data.fyre_stencils.example.stencils[0].memory
    disk        = data.fyre_stencils.example.stencils[0].disk
    owner       = data.fyre_stencils.example.stencils[0].owner
  } : null
}
