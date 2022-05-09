package tfschema

import (
	"fmt"
	"log"
	"strings"

	"github.com/env0/terratag/internal/providers"
	"github.com/env0/terratag/internal/tagging"
	"github.com/env0/terratag/internal/terraform"
	"github.com/thoas/go-funk"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfschema/tfschema"
)

var providerToClientMap = map[string]tfschema.Client{}

var customSupportedProviderNames = [...]string{"google-beta"}

func IsTaggable(dir string, resource hclwrite.Block) (bool, error) {
	var isTaggable bool
	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType) {
		providerName, _ := detectProviderName(resource)
		client, err := getClient(providerName, dir)
		if err != nil {
			return false, err
		}
		typeSchema, err := client.GetResourceTypeSchema(resourceType)
		if err != nil {
			if strings.Contains(err.Error(), "Failed to find resource type") {
				// short circuiting unfound resource due to: https://github.com/env0/terratag/issues/17
				log.Print("[WARN] Skipped ", resourceType, " as it is not YET supported")
				return false, nil
			}

			return false, err
		}

		attributes := typeSchema.Attributes
		for attribute := range attributes {
			if providers.IsTaggableByAttribute(resourceType, attribute) {
				isTaggable = true
			}
		}

		if tagging.HasResourceTagFn(resourceType) {
			isTaggable = true
		}
	}

	return isTaggable, nil
}

type TfSchemaAttribute struct {
	Name string
	Type string
}

func extractProviderNameFromResourceType(resourceType string) (string, error) {
	s := strings.SplitN(resourceType, "_", 2)
	if len(s) < 2 {
		return "", fmt.Errorf("failed to detect a provider name: %s", resourceType)
	}
	return s[0], nil
}

func detectProviderName(resource hclwrite.Block) (string, error) {
	providerAttribute := resource.Body().GetAttribute("provider")

	if providerAttribute != nil {
		providerTokens := providerAttribute.Expr().BuildTokens(hclwrite.Tokens{})
		providerName := strings.Trim(string(providerTokens.Bytes()), "\" ")
		if funk.Contains(customSupportedProviderNames, providerName) {
			return providerName, nil
		}
	}

	return extractProviderNameFromResourceType(terraform.GetResourceType(resource))
}

func getClient(providerName string, dir string) (tfschema.Client, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		// this annoyance - both tfschema and go-plugin open output
		// directly to os.Stderr, bypassing our log level filter.
		// weird to need to bypass the issue by assigning the default
		// output ¯\_(ツ)_/¯
	})
	client, exists := providerToClientMap[providerName]
	if exists {
		return client, nil
	} else {
		newClient, err := tfschema.NewClient(providerName, tfschema.Option{
			RootDir: dir,
			Logger:  logger,
		})
		if err != nil {
			return nil, err
		}

		providerToClientMap[providerName] = newClient
		return newClient, nil
	}
}
