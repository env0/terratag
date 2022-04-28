provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "a" {
  bucket = "my-another-tf-test-bucket"

  tags = "${merge("${merge(
    map("a", "b"),
    map("c", "d")
  )}", local.terratag_added_main)}"
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

