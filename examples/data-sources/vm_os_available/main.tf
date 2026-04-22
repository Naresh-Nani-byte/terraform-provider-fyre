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

# Output RedHat versions for x platform
output "x_redhat_versions" {
  description = "Available RedHat versions for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.redhat
}

# Output Ubuntu versions for x platform
output "x_ubuntu_versions" {
  description = "Available Ubuntu versions for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.ubuntu
}

# Output CentOS versions for x platform
output "x_centos_versions" {
  description = "Available CentOS versions for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.centos
}

# Output Rocky Linux versions for x platform
output "x_rocky_versions" {
  description = "Available Rocky Linux versions for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.rocky
}

# Output SLES versions for x platform
output "x_sles_versions" {
  description = "Available SLES versions for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.sles
}

# Output Windows versions for x platform
output "x_windows_versions" {
  description = "Available Windows versions for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.windows
}

# Output the default VM sizing for x platform
output "x_default_size" {
  description = "Default VM sizing constraints for x86 platform"
  value       = data.fyre_vm_os_available.x_platform.default_size
}

# Output RedHat versions for z platform
output "z_redhat_versions" {
  description = "Available RedHat versions for z platform"
  value       = data.fyre_vm_os_available.z_platform.redhat
}