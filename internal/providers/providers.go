package providers

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"strings"
	"sync"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

const AWS = "aws"
const GCP = "gcp"
const AZURE = "azure"

var resourcesToSkip = []string{"azurerm_api_management_named_value"}

var prefixes = map[string]Provider{
	"aws_":        AWS,
	"google_":     GCP,
	"azurerm_":    AZURE,
	"azurestack_": AZURE,
	"azapi_":      AZURE,
}

//go:embed azure_resource_tag_support.csv
var azureTagSupport []byte

var (
	azureTaggableResources map[string]bool
	once                   sync.Once
)

func getAzureTaggableResources() map[string]bool {
	once.Do(func() {
		reader := csv.NewReader(bytes.NewReader(azureTagSupport))
		records, err := reader.ReadAll()

		if err != nil {
			panic(err) // Or handle the error appropriately
		}

		azureTaggableResources = make(map[string]bool)
		for _, record := range records {
			azureTaggableResources[record[0]] = true
		}
	})

	return azureTaggableResources
}

func getProviderByResource(resourceType string) Provider {
	for prefix, provider := range prefixes {
		if strings.HasPrefix(resourceType, prefix) {
			return provider
		}
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

func isSimpleStringLiteral(tokens hclwrite.Tokens) (string, bool) {
	if len(tokens) == 3 &&
		tokens[0].Type == hclsyntax.TokenOQuote &&
		tokens[1].Type == hclsyntax.TokenQuotedLit &&
		tokens[2].Type == hclsyntax.TokenCQuote {
		return string(tokens[1].Bytes), true // Quoted
	}

	return "", false
}

func azureTypeIsTaggable(resource hclwrite.Block) bool {
	typeAttr := resource.Body().GetAttribute("type")

	if typeAttr == nil {
		return false
	}

	typeTokens := typeAttr.Expr().BuildTokens(nil)

	typeValue, ok := isSimpleStringLiteral(typeTokens)

	if !ok || typeValue == "" {
		return false
	}

	// split the type value to get the resource type, get everything before "@" lower cased.
	parts := strings.Split(typeValue, "@")
	if len(parts) != 2 {
		return false
	}

	resourceType := strings.ToLower(parts[0])

	// check if the resource type is taggable based on the on list of supported azure tags.
	azureTaggableResources := getAzureTaggableResources()
	if _, ok := azureTaggableResources[resourceType]; !ok {
		return false
	}

	return true
}

func IsSupportedResource(resourceType string, resource hclwrite.Block) bool {
	for _, resourceToSkip := range resourcesToSkip {
		if resourceType == resourceToSkip {
			return false
		}
	}

	if !isSupportedProvider(getProviderByResource(resourceType)) {
		return false
	}

	if strings.HasPrefix(resourceType, "azapi_") && !azureTypeIsTaggable(resource) {
		return false
	}

	return true
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
