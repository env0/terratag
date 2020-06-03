provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = {
    "Name"                    = "My bucket"
    Unquoted1                 = "I wanna be quoted"
    "AnotherName"             = "Yo"
    Unquoted2                 = "I really wanna be quoted"
    Unquoted3                 = "I really really wanna be quoted and I got a comma",
    "${max(1, 2)}"            = "Test function"
    "${local.localTagKey}"    = "Test expression"
  }
}

locals {
  localTagKey = "localTagKey"
}