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
}

locals {
  localTagKey  = "localTagKey"
  localTagKey2 = "localTagKey2"
  localMap = {
    key = "value"
  }
}

locals {
  terratag_added_main = {"some_tag"="value"}
}
