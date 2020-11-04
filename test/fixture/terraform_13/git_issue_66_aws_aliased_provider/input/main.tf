provider "aws" {
  region  = "us-east-1"
  alias   = "custom"
  version = "~> 3.0"
}

resource "aws_s3_bucket" "bucket" {
  provider = aws.custom

  bucket = "my-tf-test-bucket"
  acl    = "private"
}