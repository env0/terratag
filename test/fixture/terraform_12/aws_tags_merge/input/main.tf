provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

variable "name" {
  type    = string
  default = ""
}

variable "tags" {
  type    = map(string)
  default = {}
}

resource "aws_s3_bucket" "example" {
  bucket        = "example"
  acl           = "public-read"
  force_destroy = true

  tags = merge(
    {
      "Name" = format("%s", var.name)
    },
    var.tags
  )
}
