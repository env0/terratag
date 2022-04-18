provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = merge(tomap({
    "Name" = "My bucket"
  }), local.terratag_added_main)
}

resource "aws_s3_bucket" "c" {
  bucket = "my-tf-test-bucket2"
  acl    = "private"
  tags   = local.terratag_added_main
}

locals {
  localTagKey  = "localTagKey"
  localTagKey2 = "localTagKey2"
  localMap = {
    key = "value"
  }
}

locals {
  terratag_added_main = {"some_tag" = "value", "env0_environment_id" = "40907eff-cf7c-419a-8694-e1c6bf1d1168", "env0_project_id" = "43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}
