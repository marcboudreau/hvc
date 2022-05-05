################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# variables.tf
#   Defines the input variables for the Terraform project.
#
################################################################################

variable "target_vault_local_port" {
  description = "The NodePort of the target Vault service"
  type        = number
}

variable "source_vault_local_port" {
  description = "The NodePort of the source Vault service"
  type        = number
}

variable "kube_config_path" {
  description = "Specifies the path to the kube config file on the local system"
  type        = string
  default     = "~/.kube/config"
}

variable "kube_config_context" {
  description = "Specifies the context within the kube config file to use to for deploying the test components"
  type        = string
  default     = "docker-desktop"
}
