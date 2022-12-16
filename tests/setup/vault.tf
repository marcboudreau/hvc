################################################################################
#
# tests / setup
#   A Terraform project that sets up the integration test Vault pods and 
#   services along with Service Account and ClusterRoleBinding in the Kubernetes
#   cluster.
#
# vault.tf
#   Defines the resources to provision the target and source Vault server Pods.
#
################################################################################


locals {
  kube_namespace  = coalesce(var.kube_namespace, "default")
  vault_image_tag = coalesce(var.vault_image_tag, "latest")
  vault_container_args = [
    "server",
    "-dev",
    "-dev-listen-address=0.0.0.0:8200",
    "-dev-root-token-id=root",
  ]
}

resource "kubernetes_service_account" "vault" {
  metadata {
    name      = "vault"
    namespace = local.kube_namespace
  }
}

resource "kubernetes_cluster_role_binding" "vault_auth_delegator" {
  metadata {
    name = "target-vault-cluster-role"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "system:auth-delegator"
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.vault.metadata[0].name
    namespace = "default"
  }
}

resource "kubernetes_pod" "target_vault" {
  metadata {
    name      = "target-vault"
    namespace = local.kube_namespace
    labels = {
      app = "target-vault"
    }
  }
  spec {
    container {
      image = "vault:${local.vault_image_tag}"
      name = "target-vault"
      command = ["vault"]
      args = local.vault_container_args

      port {
        container_port = 8200
      }
    }
    restart_policy = "Never"
    service_account_name = kubernetes_service_account.vault.metadata[0].name
  }
}

resource "kubernetes_service" "target_vault" {
  metadata {
    name      = "target-vault"
    namespace = local.kube_namespace
  }
  spec {
    selector = {
      app = "target-vault"
    }
    port {
      port = 8200
      target_port = 8200
    }
    type = "NodePort"
  }
}

resource "kubernetes_pod" "source_vault" {
  metadata {
    name      = "source-vault"
    namespace = local.kube_namespace
    labels = {
      app = "source-vault"
    }
  }
  spec {
    container {
      image = "vault:${local.vault_image_tag}"
      name = "source-vault"
      command = ["vault"]
      args = local.vault_container_args

      port {
        container_port = 8200
      }
    }
    restart_policy = "Never"
    service_account_name = kubernetes_service_account.vault.metadata[0].name
  }
}

resource "kubernetes_service" "source_vault" {
  metadata {
    name      = "source-vault"
    namespace = local.kube_namespace
  }
  spec {
    selector = {
      app = "source-vault"
    }
    port {
      port = 8200
      target_port = 8200
    }
    type = "NodePort"
  }
}
