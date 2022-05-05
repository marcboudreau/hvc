################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# test1.tf
#   Defines the kubernetes resources for the test1 test case.
#
################################################################################

resource "kubernetes_namespace" "test1" {
  metadata {
    name = "test1"
  }
}

resource "kubernetes_config_map" "test1" {
  metadata {
    name = "copyjob"
    namespace = kubernetes_namespace.test1.metadata[0].name
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
      }
      "copies"=[
        {
          "mount-point"="kv"
          "path"="tc1/secret1"
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
              "source"="s1"
              "mount-point"="kv1"
              "path"="path3/secret1"
              "key"="k"
            }
          }
        }
      ]
    })
  }
}

resource "kubernetes_service_account" "test1_hvc" {
  metadata {
    name = "hvc"
    namespace = kubernetes_namespace.test1.metadata[0].name
  }
}
