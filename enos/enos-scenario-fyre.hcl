# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

scenario "fyre" {
  matrix {
    use = ["dev", "registry"]
  }

  terraform_cli = matrix.use == "dev" ? terraform_cli.dev : terraform_cli.default

  variable "vm_id" {
    type        = string
    description = "VM ID to use for testing VM data sources"
    default     = "1-8103661"
  }

  step "test_datasources" {
    module = module.datasources

    variables {
      vm_id = var.vm_id
    }
  }

  output "user" {
    value = step.test_datasources.user
  }

  output "quota_svl" {
    value = step.test_datasources.quota_svl
  }

  output "quota_rtp" {
    value = step.test_datasources.quota_rtp
  }

  output "vm_status" {
    value = step.test_datasources.vm_status
  }

  output "vm_details" {
    value = step.test_datasources.vm_details
  }

  output "vm_snapshots" {
    value = step.test_datasources.vm_snapshots
  }
}
