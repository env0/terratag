terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

provider "azapi" {
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-rg"
  location = "west europe"
  tags     = local.terratag_added_main
}

resource "azurerm_user_assigned_identity" "example" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  tags                = local.terratag_added_main
}

resource "azapi_resource" "example" {
  type      = "Microsoft.ContainerRegistry/registries@2020-11-01-preview"
  name      = "registry1"
  parent_id = azurerm_resource_group.example.id

  location = azurerm_resource_group.example.location
  identity {
    type         = "SystemAssigned, UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.example.id]
  }

  body = {
    sku = {
      name = "Standard"
    }
    properties = {
      adminUserEnabled = true
    }
  }

  tags = merge({
    "Key" = "Value"
  }, local.terratag_added_main)

  response_export_values = ["properties.loginServer", "properties.policies.quarantinePolicy.status"]
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

