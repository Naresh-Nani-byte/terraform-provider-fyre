# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

terraform_cli "default" {
}

terraform_cli "dev" {
  provider_installation {
    dev_overrides = {
      "hashicorp-forge/fyre" = abspath(joinpath(path.root, "../dist"))
    }
    direct {}
  }
}

terraform "default" {
  required_version = ">= 1.2.0"

  required_providers {
    fyre = {
      source = "registry.terraform.io/hashicorp-forge/fyre"
    }
  }
}
