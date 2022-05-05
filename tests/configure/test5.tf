################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# test5.tf
#   Defines the kubernetes resources for the test5 test case.
#
################################################################################

resource "kubernetes_namespace" "test5" {
  metadata {
    name = "test5"
  }
}

resource "kubernetes_config_map" "test5" {
  metadata {
    name = "copyjob"
    namespace = kubernetes_namespace.test5.metadata[0].name
  }

  data = {
    "spec.json" = jsonencode({
      "target"={
        "address"="http://target-vault.default:8200"
        "login"={
          "token"="invalid.token"
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
          "path"="tc5/secret1"
          "secret"={
            "source"="s1"
            "mount-point"="kv1"
            "path"="path1/secret1"
          }
        }
      ]
    })
  }
}

resource "kubernetes_service_account" "test5_hvc" {
  metadata {
    name = "hvc"
    namespace = kubernetes_namespace.test5.metadata[0].name
  }
}
