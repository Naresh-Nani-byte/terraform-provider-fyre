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

data "fyre_clusters" "example" {
}

output "clusters_info" {
  value = data.fyre_clusters.example
}
