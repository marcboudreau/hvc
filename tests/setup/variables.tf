################################################################################
#
# tests / setup
#   A Terraform project that sets up the integration test Vault pods and 
#   services along with Service Account and ClusterRoleBinding in the Kubernetes
#   cluster.
#
# variables.tf
#   Defines the input variables for the Terraform project.
#
################################################################################

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

variable "kube_namespace" {
  description = "Specifies an alternate Kubernetes namespace where the components will be deployed. If this variable is left empty, the namespace 'default' is used"
  type        = string
  default     = ""
}

variable "vault_image_tag" {
  description = "Specifies a specific Docker Image tag for the Vault image. If this variable is left empty, the tag 'latest' is used"
  type        = string
  default     = ""
}
