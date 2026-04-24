# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

module "datasources" {
  source = "./modules/datasources"
}

module "resources" {
  source = "./modules/resources"
}
