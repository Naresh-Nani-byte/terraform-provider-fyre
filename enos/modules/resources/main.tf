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

variable "site" {
  type        = string
  description = "Site location for the test VM"
  default     = "svl"
}

# Get available OS options
data "fyre_vm_os_available" "x_platform" {
  platform = "x"
  site     = var.site
}

# Create a test VM using the first available OS
resource "fyre_vm" "test" {
  os          = data.fyre_vm_os_available.x_platform.os_options[0]
  cpu         = 2
  memory      = 4
  description = "Enos test VM for terraform-provider-fyre"
  platform    = "x"
  site        = var.site

  # Set expiration to 2 hours for quick cleanup
  expiration = "2"
}

output "vm" {
  value = fyre_vm.test
}
