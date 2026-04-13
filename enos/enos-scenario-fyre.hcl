# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

scenario "fyre" {
  matrix {
    use = ["dev", "registry"]
  }

  terraform_cli = matrix.use == "dev" ? terraform_cli.dev : terraform_cli.default

  step "test_datasources" {
    module = module.datasources
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
}
