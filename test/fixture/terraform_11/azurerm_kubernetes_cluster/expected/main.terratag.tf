provider "azurerm" {
  version = "~>2.0"
  features {}
}

resource "azurerm_resource_group" "non_existing_tags_rg" {
  name     = "non_existing_tags_rg"
  location = "northeurope"
  tags     = "${local.terratag_added_main}"
}

resource "azurerm_kubernetes_cluster" "non_existing_tags_cluster" {
  name                = "non_existing_tags_cluster"
  location            = "${azurerm_resource_group.non_existing_tags_rg.location}"
  resource_group_name = "${azurerm_resource_group.non_existing_tags_rg.name}"
  dns_prefix          = "test"

  service_principal {
    client_id     = "id"
    client_secret = "secret"
  }

  default_node_pool {
    name       = "pool"
    node_count = 2
    vm_size    = "Standard_D2_v2"
    tags       = "${local.terratag_added_main}"
  }

  network_profile {
    load_balancer_sku = "Standard"
    network_plugin    = "kubenet"
  }

  tags = "${local.terratag_added_main}"
}

resource "azurerm_resource_group" "existing_tags_rg" {
  name     = "existing_tags_rg"
  location = "northeurope"
  tags     = "${local.terratag_added_main}"
}

resource "azurerm_kubernetes_cluster" "existing_tags_cluster" {
  name                = "existing_tags_cluster"
  location            = "${azurerm_resource_group.existing_tags_rg.location}"
  resource_group_name = "${azurerm_resource_group.existing_tags_rg.name}"
  dns_prefix          = "test"

  service_principal {
    client_id     = "id"
    client_secret = "secret"
  }

  default_node_pool {
    name       = "pool"
    node_count = 2
    vm_size    = "Standard_D2_v2"
    tags       = "${merge( map("existing","tag"), local.terratag_added_main)}"
  }

  network_profile {
    load_balancer_sku = "Standard"
    network_plugin    = "kubenet"
  }

  tags = "${merge( map("existing","tag"), local.terratag_added_main)}"
}




locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

