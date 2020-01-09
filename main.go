package main

import (
	"encoding/json"
	"github.com/bmatcuk/doublestar"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"github.com/thoas/go-funk"
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
	outputAsString := strings.TrimSpace(string(output))
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

	for i, match := range matches {
		resolvedMatch, err := filepath.EvalSymlinks(match)
		matches[i] = resolvedMatch
		panicOnError(err, nil)
	}
	matches = funk.UniqString(matches)

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
	log.Print("Processing file ", path)
	src, err := ioutil.ReadFile(path)
	panicOnError(err, nil)

	file, diagnostics := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if diagnostics.HasErrors() {
		hclErrors := diagnostics.Errs()
		log.Fatalln(hclErrors)
	}

	terratag := TerratagLocal{
		Found: map[string]string{},
		Added: tags,
	}

	var swappedTagsStrings []string
	for _, block := range file.Body().Blocks() {
		if block.Type() == "resource" {
			resourceType := block.Labels()[0]
			log.Print("Processing resource ", block.Labels())

			isTaggable := isTaggable(dir, resourceType)

			if isTaggable {
				swappedTagsStrings = tagResource(terratag, block, tags, swappedTagsStrings, tfVersion)
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if swappedTagsStrings != nil {

		locals := file.Body().AppendNewBlock("locals", nil)

		ctyFound := map[string]cty.Value{}
		for key, value := range terratag.Found {
			ctyFound[key] = cty.StringVal(value)
		}

		var ctyFoundMap cty.Value
		if len(ctyFound) != 0 {
			ctyFoundMap = cty.MapVal(ctyFound)
		} else {
			ctyFoundMap = cty.MapValEmpty(cty.String)
		}

		ctyTerratag := cty.ObjectVal(map[string]cty.Value{
			"added": cty.StringVal(terratag.Added),
			"found": ctyFoundMap,
		})

		locals.Body().SetAttributeValue("terratag", ctyTerratag)

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
		panic(err)
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

func tagResource(terratag TerratagLocal, resource *hclwrite.Block, tags string, swappedTagsStrings []string, tfVersion int) []string {
	log.Print("Resource taggable, processing...")

	hasExistingTags := moveExistingTags(terratag, resource)

	tagsValue := ""
	if hasExistingTags {
		tagsValue = "merge(local.terratag.found." + getResourceExistingTagsKey(resource) + ", local.terratag.added)"
	} else {
		tagsValue = "local.terratag.added"
	}

	if tfVersion == 11 {
		tagsValue = "${" + tagsValue + "}"
	}

	resource.Body().SetAttributeValue("tags", cty.StringVal(tagsValue))

	swappedTagsStrings = append(swappedTagsStrings, tagsValue)
	return swappedTagsStrings
}

func moveExistingTags(terratag TerratagLocal, resource *hclwrite.Block) bool {
	existingTags := ""

	// First we try to find tags as attribute
	tagsAttribute := resource.Body().GetAttribute("tags")

	if tagsAttribute != nil {
		// If attribute found, get its value
		log.Print("Preexisting tags ATTRIBUTE found on resource. Merging.")
		existingTags = string(tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{}).Bytes())
	} else {
		// Otherwise, we try to get tags as block
		tagsBlock := resource.Body().FirstMatchingBlock("tags", nil)
		if tagsBlock != nil {
			existingTags = getExistingTagsFromBlock(tagsBlock, existingTags)
			// If we did get tags from block, we will now remove that block, as we're going to add a merged tags ATTRIBUTE
			removeBlockResult := resource.Body().RemoveBlock(tagsBlock)
			if removeBlockResult == false {
				log.Fatal("Failed to remove found tags block!")
			}
		}
	}

	if existingTags != "" {
		terratag.Found[getResourceExistingTagsKey(resource)] = existingTags
		return true
	}
	return false
}

func getResourceExistingTagsKey(resource *hclwrite.Block) string {
	return strings.Join(resource.Labels(), "__")
}

func getExistingTagsFromBlock(tagsBlock *hclwrite.Block, existingTags string) string {
	var mapAttributes []string
	for key, attribute := range tagsBlock.Body().Attributes() {
		value := string(attribute.Expr().BuildTokens(hclwrite.Tokens{}).Bytes())
		mapAttributes = append(mapAttributes, key+"="+value)
	}

	if mapAttributes != nil {
		log.Print("Preexisting tags BLOCK found on resource. Merging.")
		existingTags = "{" + strings.Join(mapAttributes, ",") + "}"
	}
	return existingTags
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

type TerratagLocal struct {
	Found map[string]string
	Added string
}
