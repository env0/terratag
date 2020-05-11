package tfschema

import (
	"encoding/json"
	"github.com/env0/terratag/errors"
	"github.com/env0/terratag/providers"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"log"
	"os/exec"
	"strings"
)

func IsTaggable(dir string, resource hclwrite.Block) (bool, bool) {
	isTaggable := false
	isTaggableViaSpecialTagBlock := false

	if providers.IsTaggableResource(resource) {
		resourceType := resource.Labels()[0]
		command := exec.Command("tfschema", "resource", "show", "-format=json", resourceType)
		command.Dir = dir
		output, err := command.Output()
		outputAsString := string(output)

		if err != nil {
			if strings.Contains(outputAsString, "Failed to find resource type") {
				// short circuiting unfound resource due to: https://github.com/env0/terratag/issues/17
				log.Print("Skipped ", resourceType, " as it is not YET supported")
				return false, false
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

			if providers.IsTaggableByAttribute(resource, attribute.Name) {
				isTaggable = true
			}
		}

		if resourceType == "aws_autoscaling_group" {
			isTaggableViaSpecialTagBlock = true
		}
	}

	return isTaggable, isTaggableViaSpecialTagBlock
}

type TfSchemaAttribute struct {
	Name string
	Type string
}
