# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    fyre = {
      source = "hashicorp-forge/fyre"
    }

    enos = {
      source = "hashicorp-forge/enos"
    }
  }
}

variable "product_group_id" {
  type        = number
  description = "The ID of the product group that you wish to create resources in"
}

variable "public_key_path" {
  type        = string
  description = "The path to the RFC 4716 formatted public key to use for SSH access"
}

variable "private_key_path" {
  type        = string
  description = "The path to private that corresponds to the public key for SSH access"
}

provider "enos" {
  transport = {
    ssh = {
      private_key_path = var.private_key_path
    }
  }
}

provider "fyre" {
  # Username and API key can be set via FYRE_USERNAME and FYRE_API_KEY environment variables
  # or explicitly configured here:
  # username = "your-username"
  # api_key  = "your-api-key"


  # The default product_group_id to use for resources and datasources that
  # support it in their queries. When present at this level it will be inherited.
  # You can override it on a resource level by setting the product_group_id
  # attribute to a different value or null if you wish to not use it.
  # Can also be set with FYRE_PRODUCT_GROUP_ID
  product_group_id = var.product_group_id

  # Can also be set with FYRE_SITE
  site = "svl" # Default site (svl or rtp)
}

# Get available operating systems for x86 platform
data "fyre_vm_os_available" "x" {
  platform = "x"
}

# Check if a hostname is available
data "fyre_vm_check_hostname" "example" {
  hostname = "my-test-vm"
}

# Create a VM
resource "fyre_vm" "target" {
  os             = data.fyre_vm_os_available.x.centos[0]
  platform       = "x"
  cpu            = 2
  memory         = 4
  hostname       = data.fyre_vm_check_hostname.example.is_available ? data.fyre_vm_check_hostname.example.hostname : null
  description    = "Enos, Terraform, and Fyre, together at last"
  expiration     = "6" # hours
  public_network = "y" # Required to SSH from outside the datacenter
  dns            = "n" # Add the IP and hostname to DNS
  disable_delete = "n"
  quota_type     = "quick_burn" # or "product_group"
  time_to_live   = "2"          # hours, set a ttl if the quota_type is quick_burn

  # SSH pub key for access
  ssh_keys = [file(var.public_key_path)]

  # Additional disks (max 2)
  # additional_disks = ["50", "100"]
}

locals {
  // Get the public IP address
  target_ip = fyre_vm.target.ips[index(fyre_vm.target.ips.*.type, "public")]["ip"]
}

resource "enos_remote_exec" "some_remote_command" {
  depends_on = [
    fyre_vm.target
  ]

  inline = [
    // Run some commands on the remote machine via SSH
    "echo 'hello world on the target'",
    "echo 'hello stderr on the target' 1>&2",
  ]

  transport = {
    ssh = {
      host = local.target_ip
      user = "root"
    }
  }
}

# Outputs
output "target" {
  description = "Full details of the target VM"
  // The 'password' is sensitive so we have to set this if we want to see all attributes.
  // Generally you will only need output specific fields of the target resource.
  sensitive = true
  value     = fyre_vm.target
}

output "cmd_stdout" {
  value = enos_remote_exec.some_remote_command.stdout
}

output "cmd_stderr" {
  value = enos_remote_exec.some_remote_command.stderr
}
