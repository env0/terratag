package providers

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"strings"
)

func getProviderByResource(resource hclwrite.Block) Provider {
	resourceType := resource.Labels()[0]

	if strings.HasPrefix(resourceType, "aws_") {
		return "aws"
	} else if strings.HasPrefix(resourceType, "google_") {
		return "gcp"
	} else if strings.HasPrefix(resourceType, "azurerm_") || strings.HasPrefix(resourceType, "azurestack_") {
		return "azure"
	}

	return ""
}

func IsTaggableByAttribute(resource hclwrite.Block, attribute string) bool {
	provider := getProviderByResource(resource)
	tagBlockId := GetTagBlockIdByResource(resource)

	if (provider != "") && attribute == tagBlockId {
		return true
	}
	return false
}

func GetTagBlockIdByResource(resource hclwrite.Block) string {
	provider := getProviderByResource(resource)

	if provider == "aws" || provider == "azure" {
		return "tags"
	} else if provider == "gcp" {
		return "labels"
	}

	return ""
}

func IsTaggableResource(resource hclwrite.Block) bool {
	return isSupportedProvider(getProviderByResource(resource))
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
