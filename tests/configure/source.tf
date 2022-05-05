################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# source.tf
#   Defines the resources for the source Vault server.
#
################################################################################

provider "vault" {
  alias = "source"

  address = "http://localhost:${var.source_vault_local_port}"
  token   = "root"
}

resource "vault_mount" "source_kv" {
  provider = vault.source
  count    = 2

  path = "kv${count.index+1}"
  type = "kv-v2"
}

resource "vault_generic_secret" "source_kv1_path_secret1" {
  provider = vault.source
  count    = 3

  path = "${vault_mount.source_kv[0].path}/path${count.index+1}/secret1"

  data_json = <<EOT
{
  "k": "path${count.index+1}/secret1"
}
EOT
}

resource "vault_policy" "source_kv1" {
  provider = vault.source

  name = "hvc-kv1"

  policy = <<EOT
path "${vault_mount.source_kv[0].path}/metadata/*" {
  capabilities = ["read"]
}

path "${vault_mount.source_kv[0].path}/data/*" {
  capabilities = ["read"]
}
EOT
}

resource "vault_generic_secret" "source_kv2_path_secret2" {
  provider = vault.source
  count    = 3

  path = "${vault_mount.source_kv[1].path}/path${count.index+1}/secret2"

  data_json = <<EOT
{
  "k": "path${count.index+1}/secret2"
}
EOT
}

resource "vault_generic_secret" "source_kv2_path_secret1" {
  provider = vault.source

  path = "${vault_mount.source_kv[1].path}/path2/secret1"

  data_json = <<EOT
{
  "k": "path2/secret1"
}
EOT
}

resource "vault_policy" "source_kv2" {
  provider = vault.source

  name = "hvc-kv2"

  policy = <<EOT
path "${vault_mount.source_kv[1].path}/metadata/*" {
  capabilities = ["read"]
}

path "${vault_mount.source_kv[1].path}/data/*" {
  capabilities = ["read"]
}
EOT
}

resource "vault_auth_backend" "source_kubernetes" {
  provider = vault.source

  type = "kubernetes"
}

resource "vault_kubernetes_auth_backend_config" "source_kubernetes" {
  provider = vault.source

  backend                = vault_auth_backend.source_kubernetes.path
  kubernetes_host        = "https://kubernetes:443"
  disable_iss_validation = true
}

resource "vault_kubernetes_auth_backend_role" "source_kubernetes" {
  provider = vault.source
  
  backend                          = vault_auth_backend.source_kubernetes.path
  role_name                        = "test4"
  bound_service_account_names      = ["hvc"]
  bound_service_account_namespaces = ["test4"]
  token_ttl                        = 300
  token_policies                   = [vault_policy.source_kv1.name,vault_policy.source_kv2.name]
}
