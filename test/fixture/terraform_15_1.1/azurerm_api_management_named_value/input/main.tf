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
}

resource "azurerm_api_management" "api_mngmt" {
  name                = "api0"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
  publisher_name      = "publisher"
  publisher_email     = "publisher@env0.com"

  sku_name = "Developer_1"
}

resource "azurerm_api_management_named_value" "api_mngmt_named_value" {
  name                = "value0"
  resource_group_name = azurerm_resource_group.group.name
  api_management_name = azurerm_api_management.api_mngmt.name
  display_name        = "name"
  value               = "value"
}
