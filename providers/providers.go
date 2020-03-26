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
	}

	return ""
}

func SetIsTaggableByResource(resource hclwrite.Block, attribute string, isTaggable *bool) {
	provider := getProviderByResource(resource)
	tagBlockId := GetTagBlockIdByResource(resource)

	if (provider == "aws" || provider == "gcp") && attribute == tagBlockId {
		*isTaggable = true
	}
}

func GetTagBlockIdByResource(resource hclwrite.Block) string {
	provider := getProviderByResource(resource)

	if provider == "aws" {
		return "tags"
	} else if provider == "gcp" {
		return "labels"
	}

	return ""
}

func IsSupportedResource(resource hclwrite.Block) bool {
	return isSupportedProvider(getProviderByResource(resource))
}

func isSupportedProvider(provider Provider) bool {
	switch provider {
	case "aws":
		return true
	case "gcp":
		return true
	default:
		return false
	}
}

type Provider string
