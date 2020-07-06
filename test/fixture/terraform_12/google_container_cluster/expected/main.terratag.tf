resource "google_container_cluster" "no-labels-cluster" {
  name     = "cluster"
  location = "us-central1"

  node_config {
    machine_type = "n1-standard-1"
    labels       = local.terratag_added_main
  }

  node_pool {
    node_config {
      machine_type = "n1-standard-1"
      labels       = local.terratag_added_main
    }
  }
  resource_labels = local.terratag_added_main
}

resource "google_container_node_pool" "no-labels-pool" {
  cluster = google_container_cluster.no-labels-cluster.name

  node_config {
    machine_type = "n1-standard-1"
    labels       = local.terratag_added_main
  }
}

resource "google_container_cluster" "existing-labels-cluster" {
  name     = "cluster2"
  location = "us-central1"

  resource_labels = merge( map("foo" , "bar"), local.terratag_added_main)

  node_config {
    machine_type = "n1-standard-1"
    labels       = merge( map("foo" , "bar"), local.terratag_added_main)
  }

  node_pool {
    node_config {
      machine_type = "n1-standard-1"
      labels       = merge( map("foo" , "bar"), local.terratag_added_main)
    }
  }
}

resource "google_container_node_pool" "existing-labels-pool" {
  cluster = google_container_cluster.existing-labels-cluster.name

  node_config {
    machine_type = "n1-standard-1"
    labels       = merge( map("foo" , "bar"), local.terratag_added_main)
  }
}
locals {
  terratag_added_main = {"test"="test"}
}

