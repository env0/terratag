resource "google_storage_bucket" "static-site-tofu" {
  name          = "image-store-tofu.com"
  location      = "EU"
  force_destroy = true

  labels = {
    "foo" = "bar"
  }
}
