# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 1.2.0"

  required_providers {
    fyre = {
      source = "registry.terraform.io/hashicorp-forge/fyre"
    }
  }
}

variable "vm_id" {
  type        = string
  description = "VM ID to use for testing VM data sources"
}

data "fyre_quota" "rtp" {
  site = "rtp"
}

data "fyre_quota" "svl" {
  site = "svl"
}

data "fyre_user" "current" {}

data "fyre_vm_status" "test" {
  vm_id = var.vm_id
  site  = "svl"
}

data "fyre_vm_details" "test" {
  vm_id = var.vm_id
  site  = "svl"
}

data "fyre_vm_snapshots" "test" {
  vm_id = var.vm_id
  site  = "svl"
}

data "fyre_vm_os_available" "x_platform" {
  platform = "x"
  site     = "svl"
}

data "fyre_vm_os_available" "z_platform" {
  platform = "z"
  site     = "svl"
}

data "fyre_vm_check_hostname" "test" {
  hostname = data.fyre_vm_details.test.hostname
  site     = "svl"
}

data "fyre_stencils" "test" {
  product_group_id = data.fyre_user.current.development.default_product_group_id
  site             = "svl"
}

output "quota_svl" {
  value = data.fyre_quota.svl
}

output "quota_rtp" {
  value = data.fyre_quota.rtp
}

output "user" {
  value = data.fyre_user.current
}

output "vm_status" {
  value = data.fyre_vm_status.test
}

output "vm_details" {
  value = data.fyre_vm_details.test
}

output "vm_snapshots" {
  value = data.fyre_vm_snapshots.test
}

output "vm_os_available_x" {
  value = data.fyre_vm_os_available.x_platform
}

output "vm_os_available_z" {
  value = data.fyre_vm_os_available.z_platform
}

output "vm_check_hostname" {
  value = data.fyre_vm_check_hostname.test
}

output "stencils" {
  value = data.fyre_stencils.test
}
