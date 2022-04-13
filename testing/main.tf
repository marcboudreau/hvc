################################################################################
#
# hvc / testing
#   An integration testing framework.
#
# main.tf
#   Defines the terraform configuration and resources.
#
################################################################################

terraform {
  required_version = "~> 1.1.0"

  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "3.4.1"
    }
  }
}
