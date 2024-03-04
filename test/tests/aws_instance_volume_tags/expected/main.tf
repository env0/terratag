terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "instance_tags" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  tags = merge({
    "a" = "b"
  }, local.terratag_added_main)
  volume_tags = local.terratag_added_main
}

resource "aws_instance" "volume_tags" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
  }

  ebs_block_device {
    device_name = "abcdefg"
  }

  volume_tags = merge({
    "c" = "d"
  }, local.terratag_added_main)

  tags = merge({
    "a" = "b"
  }, local.terratag_added_main)
}

resource "aws_instance" "tags_in_root_block" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
    tags = merge({
      "a" = "b"
    }, local.terratag_added_main)
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags        = local.terratag_added_main
  }
  tags = local.terratag_added_main
}

resource "aws_instance" "tags_in_ebs_block" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
    tags        = local.terratag_added_main
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags = merge({
      "a" = "b"
    }, local.terratag_added_main)
  }
  tags = local.terratag_added_main
}

resource "aws_instance" "tags_in_both_blocks" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
    tags = merge({
      "c" = "d"
    }, local.terratag_added_main)
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags = merge({
      "a" = "b"
    }, local.terratag_added_main)
  }
  tags = local.terratag_added_main
}

resource "aws_instance" "multiple_tags" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  ebs_block_device {
    device_name = "abcdefg"
    tags = merge({
      "a" = "b"
    }, local.terratag_added_main)
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags = merge({
      "c" = "d"
    }, local.terratag_added_main)
  }
  tags = local.terratag_added_main
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

