provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "a" {
  bucket = "my-another-tf-test-bucket"

  tags = "${merge(map("a", "b"), map("c", "d"))}"
}
