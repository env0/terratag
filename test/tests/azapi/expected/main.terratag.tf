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

resource "azapi_resource" "example2" {
  type      = "Microsoft.ContainerRegistry/registries@2020-11-01-preview"
  name      = "registry2"
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

  response_export_values = ["properties.loginServer", "properties.policies.quarantinePolicy.status"]
  tags                   = local.terratag_added_main
}

data "azurerm_synapse_workspace" "example" {
  name                = "example-workspace"
  resource_group_name = azurerm_resource_group.example.name
}

resource "azapi_data_plane_resource" "dataset" {
  type      = "Microsoft.Synapse/workspaces/datasets@2020-12-01"
  parent_id = trimprefix(data.azurerm_synapse_workspace.example.connectivity_endpoints.dev, "https://")
  name      = "example-dataset"
  body = {
    properties = {
      type = "AzureBlob",
      typeProperties = {
        folderPath = {
          value = "@dataset().MyFolderPath"
          type  = "Expression"
        }
        fileName = {
          value = "@dataset().MyFileName"
          type  = "Expression"
        }
        format = {
          type = "TextFormat"
        }
      }
      parameters = {
        MyFolderPath = {
          type = "String"
        }
        MyFileName = {
          type = "String"
        }
      }
    }
  }
}

variable "enabled" {
  type        = bool
  default     = false
  description = "whether start the spring service"
}

resource "azurerm_spring_cloud_service" "test" {
  name                = "example-spring"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_name            = "S0"
  tags                = local.terratag_added_main
}

resource "azapi_resource_action" "start" {
  type                   = "Microsoft.AppPlatform/Spring@2022-05-01-preview"
  resource_id            = azurerm_spring_cloud_service.test.id
  action                 = "start"
  response_export_values = ["*"]

  count = var.enabled ? 1 : 0
}

resource "azapi_resource_action" "stop" {
  type                   = "Microsoft.AppPlatform/Spring@2022-05-01-preview"
  resource_id            = azurerm_spring_cloud_service.test.id
  action                 = "stop"
  response_export_values = ["*"]

  count = var.enabled ? 0 : 1
}

resource "azurerm_public_ip" "example" {
  name                = "example-ip"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  allocation_method   = "Static"
  tags                = local.terratag_added_main
}

resource "azurerm_lb" "example" {
  name                = "example-lb"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.example.id
  }
  tags = local.terratag_added_main
}

resource "azurerm_lb_nat_rule" "example" {
  resource_group_name            = azurerm_resource_group.example.name
  loadbalancer_id                = azurerm_lb.example.id
  name                           = "RDPAccess"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "PublicIPAddress"
}

resource "azapi_update_resource" "example" {
  type        = "Microsoft.Network/loadBalancers@2021-03-01"
  resource_id = azurerm_lb.example.id

  body = {
    properties = {
      inboundNatRules = [
        {
          properties = {
            idleTimeoutInMinutes = 15
          }
        }
      ]
    }
  }

  depends_on = [
    azurerm_lb_nat_rule.example,
  ]
}

resource "azapi_resource" "example4" {
  type      = "Microsoft.App/containerApps/authConfigs@2024-03-01"
  name      = "current"
  parent_id = azurerm_resource_group.example.id
  body = {
    properties = {
      globalValidation = {
        redirectToProvider          = "azureactivedirectory"
        unauthenticatedClientAction = "RedirectToLoginPage"
      }
      identityProviders = {
        azureActiveDirectory = {
          enabled           = true
          isAutoProvisioned = false
          registration = {
            clientId                = "example"
            clientSecretSettingName = "microsoft-provider-authentication-secret"
            openIdIssuer            = "https://sts.windows.net/v2.0"
          }
          validation = {
            allowedAudiences = [
              "example",
            ]
            defaultAuthorizationPolicy = {
              allowedApplications = [
                "example",
              ]
            }
          }
        }
      }
      login = {}
      platform = {
        enabled        = true
        runtimeVersion = "~2"
      }
    }
  }
}

locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

