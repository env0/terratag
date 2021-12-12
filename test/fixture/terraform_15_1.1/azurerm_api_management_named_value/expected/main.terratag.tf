terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">2.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "group" {
  name     = "group0"
  location = "West Europe"
  tags     = local.terratag_added_main
}

resource "azurerm_api_management" "api_mngmt" {
  name                = "api0"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
  publisher_name      = "publisher"
  publisher_email     = "publisher@env0.com"

  sku_name = "Developer_1"
  tags     = local.terratag_added_main
}

resource "azurerm_api_management_named_value" "api_mngmt_named_value" {
  name                = "value0"
  resource_group_name = azurerm_resource_group.group.name
  api_management_name = azurerm_api_management.api_mngmt.name
  display_name        = "name"
  value               = "value"
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

