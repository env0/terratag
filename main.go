package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	. "github.com/env0/terratag/cli"
	"github.com/env0/terratag/convert"
	. "github.com/env0/terratag/errors"
	"github.com/env0/terratag/file"
	. "github.com/env0/terratag/providers"
	"github.com/env0/terratag/tagging"
	. "github.com/env0/terratag/terraform"
	. "github.com/env0/terratag/tfschema"
	"github.com/env0/terratag/utils"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/logutils"
)

func main() {
	args, isMissingArg := InitArgs()
	if isMissingArg {
		return
	}
	initLogFiltering(args.Verbose)

	Terratag(args)
}

func initLogFiltering(verbose bool) {
	level := "INFO"
	if verbose {
		level = "DEBUG"
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(level),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	hclog.DefaultOutput = filter
}

func Terratag(args Args) {
	tfVersion := GetTerraformVersion()

	if !IsTerraformInitRun(args.Dir) {
		return
	}

	matches := GetTerraformFilePaths(args.Dir)

	counters := tagDirectoryResources(args.Dir, matches, args.Tags, args.IsSkipTerratagFiles, tfVersion, args.Rename)
	log.Print("[INFO] Summary:")
	log.Print("[INFO] Tagged ", counters.taggedResources, " resource/s (out of ", counters.totalResources, " resource/s processed)")
	log.Print("[INFO] In ", counters.taggedFiles, " file/s (out of ", counters.totalFiles, " file/s processed)")
}

func tagDirectoryResources(dir string, matches []string, tags string, isSkipTerratagFiles bool, tfVersion convert.Version, rename bool) counters {
	var total counters
	for _, path := range matches {
		if isSkipTerratagFiles && strings.HasSuffix(path, "terratag.tf") {
			log.Print("[INFO] Skipping file ", path, " as it's already tagged")
		} else {
			perFile := tagFileResources(path, dir, tags, tfVersion, rename)
			total.Add(perFile)
		}
	}
	return total
}

func tagFileResources(path string, dir string, tags string, tfVersion convert.Version, rename bool) counters {
	perFileCounters := counters{
		totalFiles: 1,
	}
	log.Print("[INFO] Processing file ", path)
	var swappedTagsStrings []string

	hcl := file.ReadHCLFile(path)
	filename := file.GetFilename(path)
	terratag := convert.TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: jsonToHclMap(tags),
	}

	for _, resource := range hcl.Body().Blocks() {
		if resource.Type() == "resource" {
			log.Print("[INFO] Processing resource ", resource.Labels())
			perFileCounters.totalResources += 1

			if IsTaggable(dir, *resource) {
				log.Print("[INFO] Resource taggable, processing...", resource.Labels())
				perFileCounters.taggedResources += 1
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
				log.Print("[INFO] Resource not taggable, skipping.", resource.Labels())
			}
		}
	}

	if len(swappedTagsStrings) > 0 {
		convert.AppendLocalsBlock(hcl, filename, terratag)

		text := string(hcl.Bytes())

		swappedTagsStrings = append(swappedTagsStrings, terratag.Added)
		text = convert.UnquoteTagsAttribute(swappedTagsStrings, text)

		file.ReplaceWithTerratagFile(path, text, rename)
		perFileCounters.taggedFiles = 1
	} else {
		log.Print("[INFO] No taggable resources found in file ", path, " - skipping")
	}
	return perFileCounters
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
