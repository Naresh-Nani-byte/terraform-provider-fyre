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

data "fyre_user_api_key" "example" {
}

output "api_key_expiration" {
  value = data.fyre_user_api_key.example.expiration
}

output "api_key_details" {
  value = data.fyre_user_api_key.example.details
}
