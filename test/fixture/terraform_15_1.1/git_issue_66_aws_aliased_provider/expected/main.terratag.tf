provider "aws" {
  region  = "us-east-1"
  alias   = "custom"
  version = "~> 3.0"
}

resource "aws_s3_bucket" "bucket" {
  provider = aws.custom

  bucket = "my-tf-test-bucket"
  acl    = "private"
  tags   = local.terratag_added_main
}
locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

