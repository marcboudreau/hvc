################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# target.tf
#   Defines the resources for the target Vault server.
#
################################################################################

provider "vault" {
  alias = "target"

  address = "http://localhost:${var.target_vault_local_port}"
  token   = "root"
}

resource "vault_mount" "target_kv" {
  provider = vault.target

  path = "kv"
  type = "kv-v2"
}

resource "vault_policy" "target_kv" {
  provider = vault.target

  name = "hvc-kv"

  policy = <<EOT
path "${vault_mount.target_kv.path}/metadata/*" {
  capabilities = ["read"]
}

path "${vault_mount.target_kv.path}/data/*" {
  capabilities = ["create","update"]
}
EOT
}

resource "vault_auth_backend" "target_kubernetes" {
  provider = vault.target

  type = "kubernetes"
}

resource "vault_kubernetes_auth_backend_config" "target_kubernetes" {
  provider = vault.target

  backend                = vault_auth_backend.target_kubernetes.path
  kubernetes_host        = "https://kubernetes:443"
  disable_iss_validation = true
}

resource "vault_kubernetes_auth_backend_role" "target_kubernetes" {
  provider = vault.target
  
  backend                          = vault_auth_backend.target_kubernetes.path
  role_name                        = "test4"
  bound_service_account_names      = ["hvc"]
  bound_service_account_namespaces = ["test4"]
  token_ttl                        = 300
  token_policies                   = [vault_policy.target_kv.name]
}
