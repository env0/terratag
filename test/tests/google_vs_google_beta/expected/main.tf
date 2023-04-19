// Google beta should be tagged. However, regular google should not be tagged.

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }
    google-beta = {
      source = "hashicorp/google-beta"
    }
  }
}

provider "google" {
  project = "my-project-id"
  region  = "us-central1"
}

provider "google-beta" {
  project = "my-project-id"
  region  = "us-central1"
}

resource "google_compute_global_address" "global_ip" {
  name          = "test"
  address_type  = "EXTERNAL"
  address       = "1.2.3.4"
  network       = null
  purpose       = null
  prefix_length = 2
  ip_version    = "IPV4"
}

resource "google_compute_global_address" "global_ip2" {
  provider      = google-beta
  name          = "test"
  address_type  = "EXTERNAL"
  address       = "1.2.3.4"
  network       = null
  purpose       = null
  prefix_length = 2
  ip_version    = "IPV4"
  labels        = local.terratag_added_main
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

