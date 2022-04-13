################################################################################
#
# hvc / testing
#   An integration testing framework.
#
# target.tf
#   Defines the resources for the target Vault server.
#
################################################################################

provider "vault" {
  alias = "target"

  address = "http://localhost:8200"
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
