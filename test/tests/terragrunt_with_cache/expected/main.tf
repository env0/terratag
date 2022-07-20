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
  labels = merge({
    "foo" = "bar"
  }, local.terratag_added_main)
}

locals {
  create_cmd_bin  = "asdas"
  destroy_cmd_bin = "fsdfds"
  gcloud_bin_path = "fsdfsd"
  wait            = "Fdfsd"
  skip_download   = true
}
locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

