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
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

