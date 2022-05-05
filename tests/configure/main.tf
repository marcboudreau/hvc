################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# main.tf
#   Defines the Terraform settings and provider configurations.
#
################################################################################

terraform {
  required_version = "~> 1.1.0"

  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "3.4.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.11.0"
    }
  }
}

provider "kubernetes" {
  config_path    = var.kube_config_path
  config_context = var.kube_config_context
}
