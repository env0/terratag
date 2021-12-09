terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
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
    join("-", ["foo", "bar"]) = "Test function"
    (local.localTagKey)       = "Test expression"
    "${local.localTagKey2}"   = "Test variable as key"
    "Yo-${local.localTagKey}" = "Test variable inside key"
    "Test variable as value" = "${local.localTagKey}"
    "Test variable inside value"  = "Yo-${local.localTagKey}"
  }
}

resource "aws_s3_bucket" "a" {
  bucket = "my-another-tf-test-bucket"
  acl    = "private"

  tags = local.localMap
}

locals {
  localTagKey = "localTagKey"
  localTagKey2 = "localTagKey2"
  localMap = {
    key = "value"
  }
}