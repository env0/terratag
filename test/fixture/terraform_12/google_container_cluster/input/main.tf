resource "google_container_cluster" "no-labels-cluster" {
  name = "cluster"
  location = "us-central1"

  node_config {
    machine_type = "n1-standard-1"
  }

  node_pool {
    node_config {
      machine_type = "n1-standard-1"
    }
  }
}

resource "google_container_node_pool" "no-labels-pool" {
  cluster = google_container_cluster.no-labels-cluster.name

  node_config {
    machine_type = "n1-standard-1"
  }
}

resource "google_container_cluster" "existing-labels-cluster" {
  name = "cluster2"
  location = "us-central1"

  resource_labels = {
    foo = "bar"
  }

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