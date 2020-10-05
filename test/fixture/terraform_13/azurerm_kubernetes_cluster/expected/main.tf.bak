provider "azurerm" {
  version = "~>2.0"
  features {}
}

resource "azurerm_resource_group" "non_existing_tags_rg" {
  name     = "non_existing_tags_rg"
  location = "northeurope"
}

resource "azurerm_kubernetes_cluster" "non_existing_tags_cluster" {
  name                = "non_existing_tags_cluster"
  location            = azurerm_resource_group.non_existing_tags_rg.location
  resource_group_name = azurerm_resource_group.non_existing_tags_rg.name
  dns_prefix          = "test"

  service_principal {
    client_id     = "id"
    client_secret = "secret"
  }

  default_node_pool {
    name       = "pool"
    node_count = 2
    vm_size    = "Standard_D2_v2"
  }

  network_profile {
    load_balancer_sku = "Standard"
    network_plugin    = "kubenet"
  }

}

resource "azurerm_resource_group" "existing_tags_rg" {
  name     = "existing_tags_rg"
  location = "northeurope"
}

resource "azurerm_kubernetes_cluster" "existing_tags_cluster" {
  name                = "existing_tags_cluster"
  location            = azurerm_resource_group.existing_tags_rg.location
  resource_group_name = azurerm_resource_group.existing_tags_rg.name
  dns_prefix          = "test"

  service_principal {
    client_id     = "id"
    client_secret = "secret"
  }

  default_node_pool {
    name       = "pool"
    node_count = 2
    vm_size    = "Standard_D2_v2"
    tags = {
      existing = "tag"
    }
  }

  network_profile {
    load_balancer_sku = "Standard"
    network_plugin    = "kubenet"
  }

  tags = {
    existing = "tag"
  }
}
