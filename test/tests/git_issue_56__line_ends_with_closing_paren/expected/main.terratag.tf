terraform {
  required_providers {
    kubernetes = {
      version = "2.1.0"
    }
    google = {
      version = "3.65.0"
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
  name   = "yuvu"
  labels = merge(local.terratag_added_main, local.terratag_added_main)
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

