package tfschema

import (
	"fmt"
	"github.com/thoas/go-funk"
	"log"
	"strings"

	"github.com/env0/terratag/errors"
	"github.com/env0/terratag/providers"
	"github.com/env0/terratag/tagging"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfschema/tfschema"
)

var providerToClientMap = map[string]tfschema.Client{}

var customSupportedProviderNames = [...]string{"google-beta"}

func IsTaggable(dir string, resource hclwrite.Block) bool {
	var isTaggable bool
	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType) {
		providerName, _ := detectProviderName(resource)
		client := getClient(providerName, dir)
		typeSchema, err := client.GetResourceTypeSchema(resourceType)
		if err != nil {
			if strings.Contains(err.Error(), "Failed to find resource type") {
				// short circuiting unfound resource due to: https://github.com/env0/terratag/issues/17
				log.Print("Skipped ", resourceType, " as it is not YET supported")
				return false
			} else {
				errors.PanicOnError(err, nil)
			}
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

	return isTaggable
}

type TfSchemaAttribute struct {
	Name string
	Type string
}

func extractProviderNameFromResourceType(resourceType string) (string, error) {
	s := strings.SplitN(resourceType, "_", 2)
	if len(s) < 2 {
		return "", fmt.Errorf("Failed to detect a provider name: %s", resourceType)
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

func getClient(providerName string, dir string) tfschema.Client {
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
		return client
	} else {
		newClient, err := tfschema.NewClient(providerName, tfschema.Option{
			RootDir: dir,
			Logger:  logger,
		})
		errors.PanicOnError(err, nil)

		providerToClientMap[providerName] = newClient
		return newClient
	}
}
