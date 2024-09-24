// Google beta should be tagged. However, regular google should not be tagged.

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.84.0"
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
}
