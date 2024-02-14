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


  tags = {
    Name = "terratag-test"
    env  = "test"
  }
}
