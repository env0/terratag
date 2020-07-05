package providers

import (
	"strings"
)

func getProviderByResource(resourceType string) Provider {
	if strings.HasPrefix(resourceType, "aws_") {
		return "aws"
	} else if strings.HasPrefix(resourceType, "google_") {
		return "gcp"
	} else if strings.HasPrefix(resourceType, "azurerm_") || strings.HasPrefix(resourceType, "azurestack_") {
		return "azure"
	}

	return ""
}

func IsTaggableByAttribute(resourceType string, attribute string) bool {
	provider := getProviderByResource(resourceType)
	tagBlockId := GetTagIdByResource(resourceType, false)

	if (provider != "") && attribute == tagBlockId {
		return true
	}
	return false
}

func GetTagIdByResource(resourceType string, isNested bool) string {
	provider := getProviderByResource(resourceType)

	if provider == "aws" || provider == "azure" {
		return "tags"
	} else if provider == "gcp" {
		if !isNested && resourceType == "google_container_cluster" {
			return "resource_labels"
		}
		return "labels"
	}

	return ""
}

func IsTaggableResource(resourceType string) bool {
	return isSupportedProvider(getProviderByResource(resourceType))
}

func isSupportedProvider(provider Provider) bool {
	switch provider {
	case "aws":
		return true
	case "gcp":
		return true
	case "azure":
		return true
	default:
		return false
	}
}

type Provider string
