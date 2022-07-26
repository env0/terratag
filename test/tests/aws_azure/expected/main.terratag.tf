terraform {
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 2.0"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "aws" {
  region = "us-east-1"
}

resource "azurerm_resource_group" "should_have_tags" {
  name     = "example-resources"
  location = "West Europe"
  tags = merge({
    "oh" = "my"
  }, local.terratag_added_main)
}

resource "azurerm_virtual_network" "should_not_have_tags" {
  name                = "example-network"
  resource_group_name = azurerm_resource_group.should_have_tags.name
  location            = azurerm_resource_group.should_have_tags.location
  address_space       = ["10.0.0.0/16"]
}

resource "aws_s3_bucket" "should_have_tags" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = merge({
    "Name" = "My bucket"
  }, local.terratag_added_main)
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}
