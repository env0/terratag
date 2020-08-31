package tfschema

import (
	"fmt"
	"github.com/env0/terratag/errors"
	"github.com/env0/terratag/providers"
	"github.com/env0/terratag/tagging"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfschema/tfschema"
	"log"
	"strings"
)

func IsTaggable(dir string, resource hclwrite.Block) bool {
	var isTaggable bool
	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType) {
		providerName, _ := detectProviderName(resourceType)
		client, err := tfschema.NewClient(providerName, dir)
		errors.PanicOnError(err, nil)
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

// shamefully copied from
// https://github.com/minamijoyo/tfschema/blob/8e65902597e0eb9ce7d5ac2b56bf948a1bf17429/command/meta.go#L20
func detectProviderName(name string) (string, error) {
	s := strings.SplitN(name, "_", 2)
	if len(s) < 2 {
		return "", fmt.Errorf("Failed to detect a provider name: %s", name)
	}
	return s[0], nil
}
