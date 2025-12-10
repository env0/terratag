terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

resource "kubernetes_deployment" "audit_server" {
  metadata {
    name = "terraform-example"
  }
  spec {
    template {
      metadata {
        labels = {
          app = "as"
        }
      }
      spec {
        container {
          image = "helloworld"
          name  = "as"
        }
      }
    }
    replicas = true ? 0 : (false ? 3 : 4)
    strategy {
      type = "Recreate"
    }
  }
}

resource "google_pubsub_topic" "audit_topic" {
  name = "yuvu"
  labels = local.terratag_added_main
}
