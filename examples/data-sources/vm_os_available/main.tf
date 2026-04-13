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

# Get available operating systems for x86 platform
data "fyre_vm_os_available" "x_platform" {
  platform = "x"
}

# Get available operating systems for z platform
data "fyre_vm_os_available" "z_platform" {
  platform = "z"
}

# Get available operating systems for x platform at RTP site
data "fyre_vm_os_available" "x_platform_rtp" {
  platform = "x"
  site     = "rtp"
}

# Output the available operating systems for x platform
output "x_operating_systems" {
  description = "Available operating systems for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.operating_systems
}

# Output the default VM sizing for x platform
output "x_default_size" {
  description = "Default VM sizing constraints for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.default_size
}

# Output all RedHat versions available for x platform
output "x_redhat_versions" {
  description = "Available RedHat versions for x86 platform"
  value       = lookup(data.fyre_vm_os_available.x_platform.operating_systems, "RedHat", [])
}

# Output the available operating systems for z platform
output "z_operating_systems" {
  description = "Available operating systems for z platform"
  value       = data.fyre_vm_os_available.z_platform.operating_systems
}
