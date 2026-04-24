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
  # Username and API key can be set via FYRE_USERNAME and FYRE_API_KEY environment variables
  # Site can be set via FYRE_SITE environment variable (defaults to 'svl')
  # Product group ID can be set via FYRE_PRODUCT_GROUP_ID environment variable
}

# Fetch current authenticated user details
data "fyre_user" "current" {}

# Output basic user information
output "user_email" {
  description = "Current user's email address"
  value       = data.fyre_user.current.email
}

output "user_full_name" {
  description = "Current user's full name"
  value       = data.fyre_user.current.full_name
}

output "user_authenticated" {
  description = "Whether the user is authenticated"
  value       = data.fyre_user.current.authenticated
}

# Output development environment details
output "user_development_id" {
  description = "User's development environment ID"
  value       = data.fyre_user.current.development.id
}

output "user_default_location" {
  description = "User's default location"
  value       = data.fyre_user.current.development.default_location
}

output "user_product_groups" {
  description = "User's product groups"
  value       = data.fyre_user.current.development.product_groups
}

output "user_roles" {
  description = "User's roles in the development environment"
  value       = data.fyre_user.current.development.roles
}

output "user_authorizations" {
  description = "User's authorizations"
  value       = data.fyre_user.current.development.authorizations
}

# Output sentry authentication details
output "user_sentry_status" {
  description = "User's sentry authentication status"
  value       = data.fyre_user.current.sentry.status
}

output "user_2fa_status" {
  description = "User's two-factor authentication status"
  value       = data.fyre_user.current.sentry.two_fa_status
}

output "user_auth_method" {
  description = "User's authentication method"
  value       = data.fyre_user.current.sentry.auth_method
}

# Example: Using user data in other resources
# This demonstrates how you might use the user data source
# to conditionally configure other resources based on user attributes

locals {
  is_admin = contains(data.fyre_user.current.development.roles, "admin")
  user_id  = data.fyre_user.current.development.id
}

output "is_admin" {
  description = "Whether the current user has admin role"
  value       = local.is_admin
}