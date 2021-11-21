terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "4.1.0"
    }
  }
}

resource "google_container_cluster" "no-labels-cluster" {
  name     = "cluster"
  location = "us-central1"

  node_config {
    machine_type = "n1-standard-1"
  }

  node_pool {
    node_config {
      machine_type = "n1-standard-1"
    }
  }
  resource_labels = local.terratag_added_main
}

resource "google_container_node_pool" "no-labels-pool" {
  cluster = google_container_cluster.no-labels-cluster.name

  node_config {
    machine_type = "n1-standard-1"
  }
}

resource "google_container_cluster" "existing-labels-cluster" {
  name     = "cluster2"
  location = "us-central1"

  resource_labels = merge(tomap({
    "foo" = "bar"
  }), local.terratag_added_main)

  node_config {
    machine_type = "n1-standard-1"
    labels = {
      foo = "bar"
    }
  }

  node_pool {
    node_config {
      machine_type = "n1-standard-1"
      labels = {
        foo = "bar"
      }
    }
  }
}

resource "google_container_node_pool" "existing-labels-pool" {
  cluster = google_container_cluster.existing-labels-cluster.name

  node_config {
    machine_type = "n1-standard-1"
    labels = {
      foo = "bar"
    }
  }
}
locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

