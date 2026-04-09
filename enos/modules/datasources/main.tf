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

data "fyre_user" "current" {}

output "user" {
  value = data.fyre_user.current
}
