provider "azurestack" {
}

resource "azurestack_resource_group" "test" {
  name     = "production"
  location = "West US"
}

resource "azurestack_virtual_network" "test" {
  name                = "production-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

}

resource "azurestack_virtual_network" "test2" {
  name                = "production-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tags = {
    "yo" = "ho"
  }
}