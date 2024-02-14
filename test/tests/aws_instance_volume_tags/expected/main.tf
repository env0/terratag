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

resource "aws_instance" "ubuntu" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"


  tags = merge({
    "Name" = "terratag-test"
    "env"  = "test"
  }, local.terratag_added_main)
  volume_tags = local.terratag_added_main
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

