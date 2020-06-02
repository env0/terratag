package main

import (
	"encoding/json"
	. "github.com/env0/terratag/cli"
	"github.com/env0/terratag/convert"
	. "github.com/env0/terratag/errors"
	"github.com/env0/terratag/file"
	. "github.com/env0/terratag/providers"
	"github.com/env0/terratag/tag_keys"
	. "github.com/env0/terratag/terraform"
	. "github.com/env0/terratag/tfschema"
	"github.com/env0/terratag/utils"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"log"
	"strings"
)

func main() {
	Terratag()
}

func Terratag() {
	tags, dir, isSkipTerratagFiles, isMissingArg := InitArgs()

	tfVersion := GetTerraformVersion()

	if isMissingArg || !IsTerraformInitRun(dir) {
		return
	}

	matches := GetTerraformFilePaths(dir)

	tagDirectoryResources(dir, matches, tags, isSkipTerratagFiles, tfVersion)
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

	keys := utils.SortObjectKeys(tagsMap)

	var mapContent []string
	for _, key := range keys {
		mapContent = append(mapContent, "\""+key+"\"="+"\""+tagsMap[key]+"\"")
	}
	return "{" + strings.Join(mapContent, ",") + "}"
}
