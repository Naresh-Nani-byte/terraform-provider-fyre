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

  variable "cluster_id" {
    type        = string
    description = "Cluster ID to use for testing cluster data sources"
    default     = "15281"
  }

  step "test_datasources" {
    module = module.datasources

    variables {
      vm_id      = var.vm_id
      cluster_id = var.cluster_id
    }
  }

  step "test_resources" {
    module = module.resources
  }

  output "vm_resource" {
    value = step.test_resources.vm
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

  output "vm_os_available_x" {
    value = step.test_datasources.vm_os_available_x
  }

  output "vm_os_available_z" {
    value = step.test_datasources.vm_os_available_z
  }

  output "vm_check_hostname" {
    value = step.test_datasources.vm_check_hostname
  }

  output "cluster_details" {
    value = step.test_datasources.cluster_details
  }

  output "cluster_details_with_vms" {
    value = step.test_datasources.cluster_details_with_vms
  }

  output "clusters" {
    value = step.test_datasources.clusters
  }

  output "stencils" {
    value = step.test_datasources.stencils
  }

  output "user_api_key" {
    value = step.test_datasources.user_api_key
  }
}
