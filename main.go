package main

import (
	"encoding/json"
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

	if isMissingArg {
		return
	}

	tagDirectoryResources(dir, tags)
}

func tagDirectoryResources(dir string, tags string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			tagFileResources(path, dir, tags)
		}
		return nil
	})
}

func initArgs() (string, string, bool) {
	var tags string
	var dir string
	isMissingArg := false

	for _, arg := range os.Args {
		tags = setFlag("tags", arg, "")
		dir = setFlag("dir", arg, ".")
	}

	if tags == "" {
		log.Println("Usage: terratag -tags='{ some_tag = \"value\" }' [-dir=\".\"]")
		isMissingArg = true
	}

	return tags, dir, isMissingArg
}

func setFlag(flag string, arg string, defaultValue string) string {
	result := defaultValue
	prefix := "-" + flag
	if strings.HasPrefix(arg, prefix) {
		result = strings.TrimPrefix(arg, prefix)
	}
	return result
}

func tagFileResources(path string, dir string, tags string) {
	src, err := ioutil.ReadFile(path)
	panicOnError(err)

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
				swappedTagsStrings = tagResource(block, tags, swappedTagsStrings)
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if swappedTagsStrings != nil {
		text := string(file.Bytes())

		text = unqouteTagsAttribute(swappedTagsStrings, text)

		replaceWithTerratagFile(path, text)
	} else {
		log.Print("No taggable resources found in file ", path, " - skipping")
	}
}

func panicOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func replaceWithTerratagFile(path string, textContent string) {
	taggedFilename := strings.TrimSuffix(path, filepath.Ext(path)) + ".terratag.tf"
	backupFilename := path + ".bak"

	log.Print("Creating file ", taggedFilename)
	taggedFileError := ioutil.WriteFile(taggedFilename, []byte(textContent), 0644)
	panicOnError(taggedFileError)

	log.Print("Renaming original file from ", path, " to ", backupFilename)
	backupFileError := os.Rename(path, backupFilename)
	panicOnError(backupFileError)
}

func unqouteTagsAttribute(swappedTagsStrings []string, text string) string {
	for _, swappedTagString := range swappedTagsStrings {
		escapedTagsString := "\"" + strings.ReplaceAll(swappedTagString, "\"", "\\\"") + "\""
		text = strings.ReplaceAll(text, escapedTagsString, swappedTagString)
	}
	return text
}

func tagResource(block *hclwrite.Block, tags string, swappedTagsStrings []string) []string {
	log.Print("Resource taggable, processing...")
	existingTags := "{}"
	tagsAttribute := block.Body().GetAttribute("tags")

	if tagsAttribute != nil {
		log.Print("Preexisting tags found on resource. Merging.")
		existingTags = string(tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{}).Bytes())
	}

	mergedTags := "merge(" + existingTags + ", " + tags + ")"
	block.Body().SetAttributeValue("tags", cty.StringVal(mergedTags))

	swappedTagsStrings = append(swappedTagsStrings, mergedTags)
	return swappedTagsStrings
}

func isTaggable(dir string, resourceType string) bool {
	command := exec.Command("tfschema", "resource", "show", "-format=json", resourceType)
	command.Dir = dir
	output, err := command.Output()
	panicOnError(err)

	var schema map[string]interface{}

	err = json.Unmarshal(output, &schema)
	panicOnError(err)

	isTaggable := false
	attributes := schema["attributes"].([]interface{})
	for _, attributeMap := range attributes {
		var attribute TfSchemaAttribute
		err := mapstructure.Decode(attributeMap, &attribute)
		panicOnError(err)

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
