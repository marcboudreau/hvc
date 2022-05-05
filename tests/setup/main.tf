################################################################################
#
# tests / setup
#   A Terraform project that sets up the integration test Vault pods and 
#   services along with Service Account and ClusterRoleBinding in the Kubernetes
#   cluster.
#
# main.tf
#   Defines the Terraform settings and provider configurations.
#
################################################################################

terraform {
  required_version = "~> 1.1.0"

  required_providers {
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
