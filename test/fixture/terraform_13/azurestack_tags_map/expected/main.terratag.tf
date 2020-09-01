terraform {
  required_providers {
    azurestack = {
      source = "hashicorp/azurestack"
    }
  }
}

resource "azurestack_resource_group" "test" {
  name     = "production"
  location = "West US"
  tags     = local.terratag_added_main
}

resource "azurestack_virtual_network" "test" {
  name                = "production-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  tags = local.terratag_added_main
}

resource "azurestack_virtual_network" "test2" {
  name                = "production-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tags                = merge( map("yo" , "ho"), local.terratag_added_main)
}
locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

