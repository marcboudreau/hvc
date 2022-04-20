# hvc - Copy Job Specification

A **Copy Job Specification** is a JSON-encoded document that can be stored in
any filepath. The document can make use of environment variable expansion by
surrounding the name of an environment variable with `${` and `}`. The expansion
will be completed before the document is run through the JSON decoder.

## `target`

Use `target` to specify the target Vault server details. Each specification must
specify a `target` section.

## `target.address`

Use `target.address` to specify the scheme, host, and port address of the target
Vault server. If this key is not provided the Vault client will be configured
using the **VAULT_ADDR** environment variable if it's set, otherwise it will
default to `https://127.0.0.1:8200` as the Vault address.

## `target.login`

Use `target.login` to specify how to obtain a Vault token for the target Vault
server.

## `target.login.token`

Use `target.login.token` to specify an existing Vault token for use with the
target Vault server.

### Example: Using Environment Variable Expansion

This example shows a specification that uses the `target.login.token` key to
specify an existing Vault token that is provided by the **TARGET_VAULT_TOKEN**
environment variable.

```json
{
  "target": {
    "address": "http://localhost:8200",
    "login": {
      "token": "${TARGET_VAULT_TOKEN}"
    }
  },
  ...
}
```

## `target.login.kubernetes`

Use `target.login.kubernetes` to specify the details for using the Kubernetes
authentication method to obtain a Vault token.

### Example: Using Kubernetes Strategy

This example shows a specification that uses the `target.login.kubernetes` key
to specify the necessary details to use the Kubernetes authentication method to
perform a login operation and obtain a valid Vault token.

```json
{
  "target": {
    "address": "http://localhost:8200",
    "login": {
      "kubernetes": {
        "mount-point": "kubernetes",
        "role": "my-role",
        "jwt-path": "/var/run/secrets/kubernetes.io/serviceaccount/token"
      }
    }
  },
  ...
}
```

## `target.login.kubernetes.mount-point`

Use `target.login.kubernetes.mount-point` to specify the path where the
Kubernetes authentication method that will be used is mounted. If this key is
not provided, the *mount-point* is assumed to be `kubernetes`.

## `target.login.kubernetes.role`

Use `target.login.kubernetes.role` to specify the backend role within the
Kubernetes authentication method to use for the login operation.

## `target.login.kubernetes.jwt-path`

Use `target.login.kubernetes.jwt-path` to specify the path on the local file
system from which the Kubernetes Service Account token is to be retrieved.  If
this key is not provided, the *jwt-path* is assumed to be
`/var/run/secrets/kubernetes.io/serviceaccount/token`.

## `sources`

The `sources` key contains a map of names to Vault server details that's used to
define every source Vault used by this specification.  At least one source Vault
must be defined.

Multiple source Vault can be used to target different Vault servers (i.e.
different `.address` values). They can also be used to target the same Vault
server with different Vault tokens (i.e. different `.login` sections).

## `sources.<source_name>`

Use the `sources.<source_name>` key to specify a source Vault server to be used
by this specification.  This key has the same sub-keys as the `target` key.

The `<source_name>` used will be referenced later on in the `copies[*].values.
<target_key>.source` key.

## `copies`

The specification consists of one or more copy operations.  Each are defined as
an element of the `copies` array.

An element of the `copies` array (or a Copy element), describes a single target
secret and its constituent key-value mappings (the target secret's data) in one
of two ways:
1. By specifying a source secret with the `copies[*].secret` key (see below)
1. By specifying a map of target secret keys to source secret values using the
`copies[*].values` key (see below)

## `copies[*].mount-point`

Use the `copies[*].mount-point` key to specify the path where the KV Secrets
Engine of the target secret is mounted. If this key is not provided, the
*mount-point* is assumed to be `kv`.

## `copies[*].path`

Use the `copies[*].path` key to specify the path of the target secret within the
KV Secrets Engine.

## `copies[*].secret`

The `copies[*].secret` key is used to define an exact copy of a source secret
into a target secret.  All of the source secret's key-value pairs will be copied
when this method is used. This key cannot be used in conjunction with the
`copies[*].values` key.

### Example: Copying a Source Secret Exactly

This example uses the `copies[*].secret` key to copy a source secret exactly
as-is to a target secret.

```json
{
  ...
  "copies": [
    {
      "mount-point": "kv",
      "path": "jenkins/deployment",
      "secret": {
        "source": "source-vault",
        "mount-point": "secret",
        "path": "jenkins/deploy"
      }
    }
  ]
}
```

The example shows the contents of the `copies` key that will copy all of the
contents of the source secret `secret/jenkins/deploy` from the source Vault
server mapped to `source-vault` in the `sources` section (not shown in the
example) to a secret `kv/jenkins/deployment` in the target Vault server.

## `copies[*].secret.source`

Use the `copies[*].secret.source` key to specify the name mapped to the
appropriate source Vault defined in the `sources` map above.

## `copies[*].secret.mount-point`

Use the `copies[*].secret.mount-point` key to specify the path where the KV
Secrets Engine of the source secret is mounted. If this key is not provided, the
*mount-point* is assumed to be `kv`.

## `copies[*].secret.path`

Use the `copies[*].secret.path` key to specify the path of the source secret
within the KV Secrets Engine. If this key is not provided, the source *path*
is assumed to be the same as the target *path*.

## `copies[*].values`

The `copies[*].values` key consists of a map of keys in the target secret to a
source Vault Value section. This key cannot be used in conjunction with the
`copies[*].secret` key.

### Example: Copying a Single Value into a Target Secret

This example shows a Copy element that specifies a single value from a source
secret to be copied into the target secret.

```json
{
  ...
  "copies": [
    {
      "mount-point": "kv",
      "path": "system/email",
      "values": {
        "password": {
          "source": "src",
          "mount-point": "secret",
          "path": "development/system/deploy-secrets",
          "key": "email-password"
        }
      }
    }
  ]
}
```

The example shows that a target secret with path `kv/system/email` will have a
single key-value mapping stored: the key is `password`; and the value is
retrieved from a source secret in the Vault mapped to `src` (not shown) with
path `secret/development/system/deploy-secrets` and from the key-value mapping
`email-password`.

## `copies[*].values.<value_name>`

Use the `copies[*].values.<value_name>` key to specify the source Vault Value section for the `<value_name>` key within the target secret.

## `copies[*].values.<value_name>.source`

Use the `copies[*].values.<value_name>.source` key to specify the name mapped to
the appropriate source Vault in the `sources` map above.

## `copies[*].values.<value_name>.mount-point`

Use the `copies[*].values.<value_name>.mount-point` key to specify the path
where the KV Secrets Engine of the tarsourceget secret is mounted. If this key
is not provided, the *mount-point* is assumed to be `kv`.

## `copies[*].values.<value_name>.path`

Use the `copies[*].values.<value_name>.path` key to specify the path of the
source secret within its KV Secrets Engine. If this key is not provided, the
source *path* is assumed to be the same as the target *path* specified in the
`copies[*].path` key.

## `copies[*].values.<value_name>.key`

Use the `copies[*].values.<value_name>.key` key to specify the key of the
key-value mapping within the source secret to copy into the `<value_name>`
mapping in the target secret. If this key is not provided, the source *key* is
assumed to be the same as the target key specified in the 
`copies[*].values.<value_name>` key.
