################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# test3.tf
#   Defines the kubernetes resources for the test3 test case.
#
################################################################################

resource "kubernetes_namespace" "test3" {
  metadata {
    name = "test3"
  }
}

resource "kubernetes_config_map" "test3" {
  metadata {
    name = "copyjob"
    namespace = kubernetes_namespace.test3.metadata[0].name
  }

  data = {
    "spec.json" = jsonencode({
      "target"={
        "address"="http://target-vault.default:8200"
        "login"={
          "token"="root"
        }
      }
      "sources"={
        "s1"={
          "address"="http://source-vault.default:8200"
          "login"={
            "token"="root"
          }
        }
        "s2"={
          "address"="http://source-vault.default:8200"
          "login"={
            "token"="root"
          }
        }
      }
      "copies"=[
        {
          "mount-point"="kv"
          "path"="tc3/secret1"
          "values"={
            "k1"={
              "source"="s1"
              "mount-point"="kv1"
              "path"="path1/secret1"
              "key"="k"
            }
            "k2"={
              "source"="s1"
              "mount-point"="kv1"
              "path"="path2/secret1"
              "key"="k"
            }
            "k3"={
              "source"="s2"
              "mount-point"="kv1"
              "path"="path3/secret1"
              "key"="k"
            }
          }
        },
        {
          "mount-point"="kv"
          "path"="tc3/secret2"
          "values"={
            "k1"={
              "source"="s1"
              "mount-point"="kv2"
              "path"="path1/secret2"
              "key"="k"
            }
            "k2"={
              "source"="s2"
              "mount-point"="kv2"
              "path"="path2/secret2"
              "key"="k"
            }
            "k3"={
              "source"="s2"
              "mount-point"="kv2"
              "path"="path3/secret2"
              "key"="k"
            }
          }
        }
      ]
    })
  }
}

resource "kubernetes_service_account" "test3_hvc" {
  metadata {
    name = "hvc"
    namespace = kubernetes_namespace.test3.metadata[0].name
  }
}
