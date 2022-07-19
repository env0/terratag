resource "google_storage_bucket" "static-site" {
  name          = "image-store.com"
  location      = "EU"
  force_destroy = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  cors {
    origin          = ["http://image-store.com"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
  labels = {
    "foo" = "bar"
  }
}

locals {
  create_cmd_bin = "asdas"
  destroy_cmd_bin = "fsdfds"
  gcloud_bin_path = "fsdfsd"
  wait = "Fdfsd"
  skip_download = true
}