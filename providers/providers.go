package providers

import (
	"strings"
)

var resourcesToSkip = []string{"azurerm_api_management_named_value"}

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
	tagBlockId := GetTagIdByResource(resourceType)

	if (provider != "") && attribute == tagBlockId {
		return true
	}
	return false
}

func GetTagIdByResource(resourceType string) string {
	provider := getProviderByResource(resourceType)

	if provider == "aws" || provider == "azure" {
		return "tags"
	} else if provider == "gcp" {
		return "labels"
	}

	return ""
}

func IsSupportedResource(resourceType string) bool {
	for _, resourceToSkip := range resourcesToSkip {
		if resourceType == resourceToSkip {
			return false
		}
	}

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
