provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

module "child-module" {
  source = "../child"
}