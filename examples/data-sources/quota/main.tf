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

# Example: Get quota information for SVL site
data "fyre_quota" "svl" {
  site = "svl"
}

# Example: Get quota information for RTP site
data "fyre_quota" "rtp" {
  site = "rtp"
}

output "svl_quota" {
  description = "Quota information for SVL site"
  value       = data.fyre_quota.svl
}

output "rtp_quota" {
  description = "Quota information for RTP site"
  value       = data.fyre_quota.rtp
}

# Example: Access specific quota details
output "svl_x_platform_cpu" {
  description = "CPU quota for x platform in SVL"
  value       = data.fyre_quota.svl.details[0].x.cpu
}
