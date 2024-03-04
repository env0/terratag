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

  tags = {
    a = "b"
  }
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

  volume_tags = {
    c = "d"
  }

  tags = {
    a = "b"
  }
}

resource "aws_instance" "tags_in_root_block" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
    tags = {
      a = "b"
    }
  }

  ebs_block_device {
    device_name = "abcdefg"
  }
}

resource "aws_instance" "tags_in_ebs_block" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags = {
      a = "b"
    }
  }
}

resource "aws_instance" "tags_in_both_blocks" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  root_block_device {
    volume_size = 8
    tags = {
      c = "d"
    }
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags = {
      a = "b"
    }
  }
}

resource "aws_instance" "multiple_tags" {
  ami               = "dasdasD"
  instance_type     = "t3.micro"
  availability_zone = "us-west-2"

  ebs_block_device {
    device_name = "abcdefg"
    tags = {
      a = "b"
    }
  }

  ebs_block_device {
    device_name = "abcdefg"
    tags = {
      c = "d"
    }
  }
}
