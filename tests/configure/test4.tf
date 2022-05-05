################################################################################
#
# tests / configure
#   A Terraform project that configures the integration test components. The
#   Vault configuration for both the source and target and the Kubernetes
#   ConfigMap for the test cases.
#
# test4.tf
#   Defines the kubernetes resources for the test4 test case.
#
################################################################################

resource "kubernetes_namespace" "test4" {
  metadata {
    name = "test4"
  }
}

resource "kubernetes_config_map" "test4" {
  metadata {
    name = "copyjob"
    namespace = kubernetes_namespace.test4.metadata[0].name
  }

  data = {
    "spec.json" = jsonencode({
      "target"={
        "address"="http://target-vault.default:8200"
        "login"={
          "kubernetes"={
            "role"="test4"
          }
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
          "path"="tc4/secret1"
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

resource "kubernetes_service_account" "test4_hvc" {
  metadata {
    name = "hvc"
    namespace = kubernetes_namespace.test4.metadata[0].name
  }
}
