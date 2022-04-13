################################################################################
#
# hvc / testing
#   An integration testing framework.
#
# source.tf
#   Defines the resources for the source Vault server.
#
################################################################################

provider "vault" {
  alias = "source"

  address = "http://localhost:8300"
  token   = "root"
}

resource "vault_mount" "source_kv1" {
  provider = vault.source

  path = "kv1"
  type = "kv-v2"
}

resource "vault_generic_secret" "source_kv1_path1_secret1" {
  provider = vault.source

  path = "${vault_mount.source_kv1.path}/path1/secret1"

  data_json = <<EOT
{
  "k1": "secret1-v1",
  "k2": "secret1-v2",
  "k3": "secret1-v3"
}
EOT
}

resource "vault_policy" "source_kv1" {
  provider = vault.source

  name = "hvc-kv1"

  policy = <<EOT
path "${vault_mount.source_kv1.path}/metadata/*" {
  capabilities = ["read"]
}

path "${vault_mount.source_kv1.path}/data/*" {
  capabilities = ["read"]
}
EOT
}

resource "vault_mount" "source_kv2" {
  provider = vault.source

  path = "kv2"
  type = "kv-v2"
}

resource "vault_generic_secret" "source_kv2_path2_secret2" {
  provider = vault.source

  path = "${vault_mount.source_kv2.path}/path2/secret2"

  data_json = <<EOT
{
  "k1": "secret2-v1",
  "k2": "secret2-v2",
  "k3": "secret2-v3"
}
EOT
}

resource "vault_generic_secret" "source_kv2_path3_secret3" {
  provider = vault.source

  path = "${vault_mount.source_kv2.path}/path3/secret3"

  data_json = <<EOT
{
  "k1": "secret3-v1",
  "k2": "secret3-v2",
  "k3": "secret3-v3"
}
EOT
}

resource "vault_policy" "source_kv2" {
  provider = vault.source

  name = "hvc-kv2"

  policy = <<EOT
path "${vault_mount.source_kv2.path}/metadata/*" {
  capabilities = ["read"]
}

path "${vault_mount.source_kv2.path}/data/*" {
  capabilities = ["read"]
}
EOT
}