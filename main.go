package main

import (
	"encoding/json"
	. "github.com/env0/terratag/cli"
	"github.com/env0/terratag/convert"
	. "github.com/env0/terratag/errors"
	"github.com/env0/terratag/file"
	. "github.com/env0/terratag/providers"
	"github.com/env0/terratag/tagging"
	. "github.com/env0/terratag/terraform"
	. "github.com/env0/terratag/tfschema"
	"github.com/env0/terratag/utils"
	"github.com/hashicorp/hcl/v2/hclwrite"
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
	var swappedTagsStrings []string

	hcl := file.ReadHCLFile(path)
	filename := file.GetFilename(path)
	terratag := convert.TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: jsonToHclMap(tags),
	}

	for _, resource := range hcl.Body().Blocks() {
		if resource.Type() == "resource" {
			log.Print("Processing resource ", resource.Labels())

			if IsTaggable(dir, *resource) {
				log.Print("Resource taggable, processing...")
				result := tagging.TagResource(tagging.TagBlockArgs{
					Filename:  filename,
					Block:     resource,
					Tags:      tags,
					Terratag:  terratag,
					TagId:     GetTagIdByResource(GetResourceType(*resource)),
					TfVersion: tfVersion,
				})

				swappedTagsStrings = append(swappedTagsStrings, result.SwappedTagsStrings...)
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if len(swappedTagsStrings) > 0 {
		convert.AppendLocalsBlock(hcl, filename, terratag)

		text := string(hcl.Bytes())

		swappedTagsStrings = append(swappedTagsStrings, terratag.Added)
		text = convert.UnquoteTagsAttribute(swappedTagsStrings, text)

		file.ReplaceWithTerratagFile(path, text)
	} else {
		log.Print("No taggable resources found in file ", path, " - skipping")
	}
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
