package main

import (
	"encoding/json"
	"github.com/bmatcuk/doublestar"
	. "github.com/env0/terratag/cli"
	"github.com/env0/terratag/convert"
	. "github.com/env0/terratag/errors"
	"github.com/env0/terratag/file"
	. "github.com/env0/terratag/providers"
	"github.com/env0/terratag/tag_keys"
	. "github.com/env0/terratag/terraform"
	. "github.com/env0/terratag/tfschema"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
	"github.com/zclconf/go-cty/cty"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	tags, dir, isSkipTerratagFiles, isMissingArg := InitArgs()

	tfVersion := GetTeraformVersion()

	if isMissingArg || !isTerraformInitRun(dir) {
		return
	}

	matches := getTerraformFilePaths(dir)

	tagDirectoryResources(dir, matches, tags, isSkipTerratagFiles, tfVersion)
}

func isTerraformInitRun(dir string) bool {
	_, err := os.Stat(dir + "/.terraform")

	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalln("terraform init must run before running terratag")
			return false
		}

		message := "couldn't determine if terraform init has run"
		PanicOnError(err, &message)
	}

	return true
}

func getModulesDirsPaths(dir string) []string {
	var paths []string
	var modulesJson TerraformModulesJson

	jsonFile, err := os.Open(dir + "/.terraform/modules/modules.json")

	if os.IsNotExist(err) {
		closeErr := jsonFile.Close()
		PanicOnError(closeErr, nil)

		return paths
	}
	PanicOnError(err, nil)

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &modulesJson)
	PanicOnError(err, nil)

	for _, module := range modulesJson.Modules {
		modulePath, err := filepath.EvalSymlinks(dir + "/" + module.Dir)
		PanicOnError(err, nil)

		paths = append(paths, modulePath)
	}

	err = jsonFile.Close()
	PanicOnError(err, nil)

	return paths
}

func getTerraformFilePaths(rootDir string) []string {
	const tfFileMatcher = "/**/*.tf"

	tfFiles, err := doublestar.Glob(rootDir + tfFileMatcher)
	PanicOnError(err, nil)

	modulesDirs := getModulesDirsPaths(rootDir)

	for _, moduleDir := range modulesDirs {
		matches, err := doublestar.Glob(moduleDir + tfFileMatcher)
		PanicOnError(err, nil)

		tfFiles = append(tfFiles, matches...)
	}

	for i, tfFile := range tfFiles {
		resolvedTfFile, err := filepath.EvalSymlinks(tfFile)
		PanicOnError(err, nil)

		tfFiles[i] = resolvedTfFile
	}

	return funk.UniqString(tfFiles)
}

func tagDirectoryResources(dir string, matches []string, tags string, isSkipTerratagFiles bool, tfVersion int) {
	for _, path := range matches {
		if isSkipTerratagFiles && strings.HasSuffix(path, "terratag.tf") {
			log.Print("Skipping file ", path, " as it's already tagged")
		} else {
			tagFileResources(path, dir, tags, tfVersion)
		}
	}
}

func tagFileResources(path string, dir string, tags string, tfVersion int) {
	log.Print("Processing file ", path)
	hcl := file.ReadHCLFile(path)

	terratag := convert.TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: jsonToHclMap(tags),
	}

	filename := file.GetFilename(path)

	anyTagged := false
	var swappedTagsStrings []string

	for _, resource := range hcl.Body().Blocks() {
		if resource.Type() == "resource" {
			log.Print("Processing resource ", resource.Labels())

			isTaggable, isTaggableViaSpecialTagBlock := IsTaggable(dir, *resource)

			if isTaggable {
				if !isTaggableViaSpecialTagBlock {
					// for now, we count on it that if there's a single "tag" in the schema (unlike "tags" block),
					// then no "tags" interpolation is used, but rather multiple instances of a "tag" block
					// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html
					swappedTagsStrings = append(swappedTagsStrings, tagResource(filename, terratag, resource, tfVersion))
				} else {
					convert.AppendTagBlocks(resource, tags)
				}
				anyTagged = true
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if anyTagged {
		convert.AppendLocalsBlock(hcl, filename, terratag)

		text := string(hcl.Bytes())

		swappedTagsStrings = append(swappedTagsStrings, terratag.Added)
		text = convert.UnquoteTagsAttribute(swappedTagsStrings, text)

		file.ReplaceWithTerratagFile(path, text)
	} else {
		log.Print("No taggable resources found in file ", path, " - skipping")
	}
}

func tagResource(filename string, terratag convert.TerratagLocal, resource *hclwrite.Block, tfVersion int) string {
	log.Print("Resource taggable, processing...")

	hasExistingTags := convert.MoveExistingTags(filename, terratag, resource)

	tagsValue := ""
	if hasExistingTags {
		tagsValue = "merge( " + convert.GetExistingTagsExpression(terratag.Found[tag_keys.GetResourceExistingTagsKey(filename, resource)]) + ", local." + tag_keys.GetTerratagAddedKey(filename) + ")"
	} else {
		tagsValue = "local." + tag_keys.GetTerratagAddedKey(filename)
	}

	if tfVersion == 11 {
		tagsValue = "${" + tagsValue + "}"
	}

	resource.Body().SetAttributeValue(GetTagBlockIdByResource(*resource), cty.StringVal(tagsValue))

	return tagsValue
}

func jsonToHclMap(tags string) string {
	var tagsMap map[string]string
	err := json.Unmarshal([]byte(tags), &tagsMap)
	PanicOnError(err, nil)

	var mapContent []string
	for key, value := range tagsMap {
		mapContent = append(mapContent, "\""+key+"\"="+"\""+value+"\"")
	}
	return "{" + strings.Join(mapContent, ",") + "}"
}

type TerraformModulesJson struct {
	Modules []TerraformModuleMetadata `json:"Modules"`
}

type TerraformModuleMetadata struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}
