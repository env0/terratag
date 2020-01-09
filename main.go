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

	_, filename := filepath.Split(path)
	filename = strings.TrimSuffix(filename, filepath.Ext(path))
	filename = strings.ReplaceAll(filename, ".", "-")

	file, diagnostics := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if diagnostics.HasErrors() {
		hclErrors := diagnostics.Errs()
		log.Fatalln(hclErrors)
	}

	terratag := TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: tags,
	}

	anyTagged := false
	var swappedTagsStrings []string

	for _, block := range file.Body().Blocks() {
		if block.Type() == "resource" {
			resourceType := block.Labels()[0]
			log.Print("Processing resource ", block.Labels())

			isTaggable := isTaggable(dir, resourceType)

			if isTaggable {
				swappedTagsStrings = append(swappedTagsStrings, tagResource(filename, terratag, block, tfVersion))
				anyTagged = true
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if anyTagged {
		file.Body().AppendNewline()
		locals := file.Body().AppendNewBlock("locals", nil)
		file.Body().AppendNewline()

		for key, tokens := range terratag.Found {
			locals.Body().SetAttributeRaw(key, tokens)
		}
		locals.Body().SetAttributeValue(getTerratagAddedKey(filename), cty.StringVal(terratag.Added))

		text := string(file.Bytes())

		swappedTagsStrings = append(swappedTagsStrings, terratag.Added)
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

func tagResource(filename string, terratag TerratagLocal, resource *hclwrite.Block, tfVersion int) string {
	log.Print("Resource taggable, processing...")

	hasExistingTags := moveExistingTags(filename, terratag, resource)

	tagsValue := ""
	if hasExistingTags {
		tagsValue = "merge(local." + getResourceExistingTagsKey(filename, resource) + ", local." + getTerratagAddedKey(filename) + ")"
	} else {
		tagsValue = "local." + getTerratagAddedKey(filename)
	}

	if tfVersion == 11 {
		tagsValue = "${" + tagsValue + "}"
	}

	resource.Body().SetAttributeValue("tags", cty.StringVal(tagsValue))

	return tagsValue
}

func moveExistingTags(filename string, terratag TerratagLocal, resource *hclwrite.Block) bool {
	var existingTags hclwrite.Tokens

	// First we try to find tags as attribute
	tagsAttribute := resource.Body().GetAttribute("tags")

	if tagsAttribute != nil {
		// If attribute found, get its value
		log.Print("Preexisting tags ATTRIBUTE found on resource. Merging.")
		existingTags = tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{})
	} else {
		// Otherwise, we try to get tags as block
		tagsBlock := resource.Body().FirstMatchingBlock("tags", nil)
		if tagsBlock != nil {
			existingTags = funk.Tail(tagsBlock.BuildTokens(hclwrite.Tokens{})).(hclwrite.Tokens)
			// If we did get tags from block, we will now remove that block, as we're going to add a merged tags ATTRIBUTE
			removeBlockResult := resource.Body().RemoveBlock(tagsBlock)
			if removeBlockResult == false {
				log.Fatal("Failed to remove found tags block!")
			}
		}
	}

	if existingTags != nil {
		terratag.Found[getResourceExistingTagsKey(filename, resource)] = existingTags
		return true
	}
	return false
}

func getTerratagAddedKey(filname string) string {
	return "terratag_added_" + filname
}

func getResourceExistingTagsKey(filename string, resource *hclwrite.Block) string {
	delimiter := "__"
	return "terratag_found_" + filename + delimiter + strings.Join(resource.Labels(), delimiter)
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
	Found map[string]hclwrite.Tokens
	Added string
}
