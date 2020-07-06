package tfschema

import (
	"encoding/json"
	"github.com/env0/terratag/errors"
	"github.com/env0/terratag/providers"
	"github.com/env0/terratag/resources"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"log"
	"os/exec"
	"strings"
)

func IsTaggable(dir string, resource hclwrite.Block) bool {
	resourceType := terraform.GetResourceType(resource)

	// short-circuit if resource type is known to have a custom tagging function
	if resources.HasResourceTagFunc(resourceType) {
		return true
	}

	if providers.IsTaggableResource(resourceType) {
		command := exec.Command("tfschema", "resource", "show", "-format=json", resourceType)
		command.Dir = dir
		output, err := command.Output()
		outputAsString := string(output)

		if err != nil {
			if strings.Contains(outputAsString, "Failed to find resource type") {
				// short circuiting unfound resource due to: https://github.com/env0/terratag/issues/17
				log.Print("Skipped ", resourceType, " as it is not YET supported")
				return false
			} else {
				errors.PanicOnError(err, &outputAsString)
			}
		}

		var schema map[string]interface{}

		err = json.Unmarshal(output, &schema)
		errors.PanicOnError(err, nil)

		attributes := schema["attributes"].([]interface{})
		for _, attributeMap := range attributes {
			var attribute TfSchemaAttribute
			err := mapstructure.Decode(attributeMap, &attribute)
			errors.PanicOnError(err, nil)

			if providers.IsTaggableByAttribute(resourceType, attribute.Name) {
				return true
			}
		}
	}

	return false
}

type TfSchemaAttribute struct {
	Name string
	Type string
}
