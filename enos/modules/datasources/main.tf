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

data "fyre_quota" "rtp" {
  site = "rtp"
}

data "fyre_quota" "svl" {
  site = "svl"
}

data "fyre_user" "current" {}

output "quota_svl" {
  value = data.fyre_quota.svl
}

output "quota_rtp" {
  value = data.fyre_quota.rtp
}

output "user" {
  value = data.fyre_user.current
}
