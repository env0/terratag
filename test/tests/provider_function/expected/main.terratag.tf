terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.68.0"
    }
  }
}

resource "aws_ecr_repository" "hashicups" {
  name = "hashicups"

  image_scanning_configuration {
    scan_on_push = true
  }
  tags = local.terratag_added_main
}

output "hashicups_ecr_repository_account_id" {
  value = provider::aws::arn_parse(aws_ecr_repository.hashicups.arn).account_id
}


locals {
  terratag_added_main = {"a"="b"}
}