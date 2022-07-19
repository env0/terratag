generate "provider" {
  path = "provider.tf"
  if_exists = "overwrite_terragrunt"

  contents = <<EOF
terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }
  }
}

provider "google" {
  project     = "my-project-id"
  region      = "us-central1"
}
EOF
}


terraform {
  source = "tfr:///terraform-google-modules/gcloud/google?version=2.1.0"
}