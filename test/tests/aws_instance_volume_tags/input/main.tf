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

resource "aws_instance" "ubuntu" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  tags = {
    Name = "terratag-test"
    env  = "test"
  }
}

resource "aws_instance" "ubuntu2" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
    tags = {
      "a" = "b"
    }
  }
}
