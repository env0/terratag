package tfschema

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"os/exec"
)

func IsTaggable(dir string, resourceType string) (bool, bool) {
	command := exec.Command("tfschema", "resource", "show", "-format=json", resourceType)
	command.Dir = dir
	output, err := command.Output()
	outputAsString := string(output)
	error.PanicOnError(err, &outputAsString)

	var schema map[string]interface{}

	err = json.Unmarshal(output, &schema)
	main2.panicOnError(err, nil)

	isTaggable := false
	isTaggableViaSpecialTagBlock := false

	attributes := schema["attributes"].([]interface{})
	for _, attributeMap := range attributes {
		var attribute TfSchemaAttribute
		err := mapstructure.Decode(attributeMap, &attribute)
		main2.panicOnError(err, nil)

		if attribute.Name == "tags" {
			isTaggable = true
		}
	}

	if resourceType == "aws_autoscaling_group" {
		isTaggableViaSpecialTagBlock = true
	}

	return isTaggable, isTaggableViaSpecialTagBlock
}

type TfSchemaAttribute struct {
	Name string
	Type string
}
