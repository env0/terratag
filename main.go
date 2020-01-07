package main

import (
	"encoding/json"
	"github.com/bmatcuk/doublestar"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"github.com/zclconf/go-cty/cty"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	tags, dir, isMissingArg := initArgs()

	tfVersion := getTeraformVersion()

	if isMissingArg {
		return
	}

	tagDirectoryResources(dir, tags, tfVersion)
}

func getTeraformVersion() int {
	output, err := exec.Command("terraform", "version").Output()
	outputAsString := string(output)
	panicOnError(err, &outputAsString)

	if strings.HasPrefix(outputAsString, "Terraform v0.11") {
		return 11
	} else if strings.HasPrefix(outputAsString, "Terraform v0.12") {
		return 12
	}

	log.Fatalln("Terratag only supports Terraform 0.11.x and 0.12.x - your version says ", outputAsString)
	return -1
}

func tagDirectoryResources(dir string, tags string, tfVersion int) {
	matches, err := doublestar.Glob(dir + "/**/*.tf")
	panicOnError(err, nil)
	for _, path := range matches {
		tagFileResources(path, dir, tags, tfVersion)
	}
}

func initArgs() (string, string, bool) {
	var tags string
	var dir string
	isMissingArg := false

	tags = setFlag("tags", "")
	dir = setFlag("dir", ".")

	if tags == "" {
		log.Println("Usage: terratag -tags='{ some_tag = \"value\" }' [-dir=\".\"]")
		isMissingArg = true
	}

	return tags, dir, isMissingArg
}

func setFlag(flag string, defaultValue string) string {
	result := defaultValue
	prefix := "-" + flag + "="
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			result = strings.TrimPrefix(arg, prefix)
		}
	}

	return result
}

func tagFileResources(path string, dir string, tags string, tfVersion int) {
	src, err := ioutil.ReadFile(path)
	panicOnError(err, nil)

	file, diagnostics := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if diagnostics.HasErrors() {
		hclErrors := diagnostics.Errs()
		log.Fatalln(hclErrors)
	}

	var swappedTagsStrings []string
	for _, block := range file.Body().Blocks() {
		if block.Type() == "resource" {
			resourceType := block.Labels()[0]
			log.Print("Processing resource ", block.Labels())

			isTaggable := isTaggable(dir, resourceType)

			if isTaggable {
				swappedTagsStrings = tagResource(block, tags, swappedTagsStrings, tfVersion)
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if swappedTagsStrings != nil {
		text := string(file.Bytes())

		text = unquoteTagsAttribute(swappedTagsStrings, text)

		replaceWithTerratagFile(path, text)
	} else {
		log.Print("No taggable resources found in file ", path, " - skipping")
	}
}

func panicOnError(err error, moreInfo *string) {
	if err != nil {
		if moreInfo != nil {
			log.Println(*moreInfo)
		}
		log.Fatalln(err)
	}
}

func replaceWithTerratagFile(path string, textContent string) {
	taggedFilename := strings.TrimSuffix(path, filepath.Ext(path)) + ".terratag.tf"
	backupFilename := path + ".bak"

	log.Print("Creating file ", taggedFilename)
	taggedFileError := ioutil.WriteFile(taggedFilename, []byte(textContent), 0644)
	panicOnError(taggedFileError, nil)

	log.Print("Renaming original file from ", path, " to ", backupFilename)
	backupFileError := os.Rename(path, backupFilename)
	panicOnError(backupFileError, nil)
}

func unquoteTagsAttribute(swappedTagsStrings []string, text string) string {
	for _, swappedTagString := range swappedTagsStrings {
		escapedByWriter := strings.ReplaceAll(swappedTagString, "\"", "\\\"")

		if strings.HasPrefix(swappedTagString, "${") && strings.HasSuffix(swappedTagString, "}") {
			escapedByWriter = strings.ReplaceAll(escapedByWriter, "${", "$${")
		} else {
			escapedByWriter = "\"" + escapedByWriter + "\""
		}

		text = strings.ReplaceAll(text, escapedByWriter, swappedTagString)
	}
	return text
}

func tagResource(block *hclwrite.Block, tags string, swappedTagsStrings []string, tfVersion int) []string {
	log.Print("Resource taggable, processing...")

	tagsAttribute := block.Body().GetAttribute("tags")

	mergedTags := mergeTags(tagsAttribute, tags, tfVersion)

	block.Body().SetAttributeValue("tags", cty.StringVal(mergedTags))

	swappedTagsStrings = append(swappedTagsStrings, mergedTags)
	return swappedTagsStrings
}

func mergeTags(tagsAttribute *hclwrite.Attribute, tags string, tfVersion int) string {
	existingTags := ""
	mergedTags := ""

	if tagsAttribute != nil {
		log.Print("Preexisting tags found on resource. Merging.")
		existingTags = string(tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{}).Bytes())
	}

	switch tfVersion {
	case 11:
		if existingTags == "" {
			existingTags = "map()"
		}
		existingTags = convertExistingTagsToFunctionParameter(existingTags)

		mergedTags = "${ merge(" + existingTags + ", " + convertHclMapToHclMapFunction(tags) + " ) }"
		break
	case 12:
		if existingTags == "" {
			existingTags = "{}"
		}

		mergedTags = "merge(" + existingTags + ", " + tags + ")"
		break
	}

	return mergedTags
}

func convertExistingTagsToFunctionParameter(existingTags string) string {
	existingTags = strings.TrimSpace(existingTags)
	if isHcl1Placeholder(existingTags) {
		existingTags = strings.TrimSuffix(strings.TrimPrefix(existingTags, "\"${"), "}\"")
	} else if isHclMap(existingTags) {
		existingTags = convertHclMapToHclMapFunction(existingTags)
	}
	return existingTags
}

func convertHclMapToHclMapFunction(hclMap string) string {
	hclMapFunction := "map("

	hclMap = strings.TrimPrefix(hclMap, "{")
	hclMap = strings.TrimSuffix(hclMap, "}")

	keyValues := strings.Split(hclMap, ",")

	var functionVariables []string

	for _, keyValue := range keyValues {
		pair := strings.Split(keyValue, "=")
		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])

		if !(strings.HasPrefix(key, "\"") && strings.HasSuffix(key, "\"")) {
			key = "\"" + key + "\""
		}

		functionVariables = append(functionVariables, key)
		functionVariables = append(functionVariables, value)
	}

	hclMapFunction = hclMapFunction + strings.Join(functionVariables, ", ") + ")"

	return hclMapFunction
}

func isHclMap(existingTags string) bool {
	return strings.HasPrefix(existingTags, "{") && strings.HasSuffix(existingTags, "}")
}

func isHcl1Placeholder(existingTags string) bool {
	return strings.HasPrefix(existingTags, "\"${") && strings.HasSuffix(existingTags, "}\"")
}

func isTaggable(dir string, resourceType string) bool {
	command := exec.Command("tfschema", "resource", "show", "-format=json", resourceType)
	command.Dir = dir
	output, err := command.Output()
	outputAsString := string(output)
	panicOnError(err, &outputAsString)

	var schema map[string]interface{}

	err = json.Unmarshal(output, &schema)
	panicOnError(err, nil)

	isTaggable := false
	attributes := schema["attributes"].([]interface{})
	for _, attributeMap := range attributes {
		var attribute TfSchemaAttribute
		err := mapstructure.Decode(attributeMap, &attribute)
		panicOnError(err, nil)

		if attribute.Name == "tags" {
			isTaggable = true
		}
	}
	return isTaggable
}

type TfSchemaAttribute struct {
	Name string
	Type string
}
