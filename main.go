package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	. "github.com/env0/terratag/cli"
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/errors"
	"github.com/env0/terratag/file"
	. "github.com/env0/terratag/providers"
	"github.com/env0/terratag/tag_keys"
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

	counters := tagDirectoryResources(args.Dir, args.Filter, matches, args.Tags, args.IsSkipTerratagFiles, tfVersion, args.Rename)
	log.Print("[INFO] Summary:")
	log.Print("[INFO] Tagged ", counters.taggedResources, " resource/s (out of ", counters.totalResources, " resource/s processed)")
	log.Print("[INFO] In ", counters.taggedFiles, " file/s (out of ", counters.totalFiles, " file/s processed)")
}

func tagDirectoryResources(dir string, filter string, matches []string, tags string, isSkipTerratagFiles bool, tfVersion convert.Version, rename bool) counters {
	var total counters
	for _, path := range matches {
		if isSkipTerratagFiles && strings.HasSuffix(path, "terratag.tf") {
			log.Print("[INFO] Skipping file ", path, " as it's already tagged")
		} else {
			perFile := tagFileResources(path, dir, filter, tags, tfVersion, rename)
			total.Add(perFile)
		}
	}
	return total
}

func tagFileResources(path string, dir string, filter string, tags string, tfVersion convert.Version, rename bool) counters {
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
		switch resource.Type() {
		case "resource":
			log.Print("[INFO] Processing resource ", resource.Labels())
			perFileCounters.totalResources += 1

			matched, err := regexp.MatchString(filter, resource.Labels()[0])
			if err != nil {
				errors.PanicOnError(err, nil)
			}
			if !matched {
				log.Print("[INFO] Resource excluded by filter, skipping.", resource.Labels())
				continue
			}

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
		case "locals":
			// Checks if terratag_added_* exists.
			// If it exists no need to append it again to Terratag file.
			// Instead should override it.
			attributes := resource.Body().Attributes()
			key := tag_keys.GetTerratagAddedKey(filename)
			for attributeKey, attribute := range attributes {
				if attributeKey == key {
					mergedAdded, err := convert.MergeTerratagLocals(attribute, terratag.Added)
					if err != nil {
						errors.PanicOnError(err, nil)
					}
					terratag.Added = mergedAdded

					break
				}
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
	if err != nil {
		log.Printf("[ERROR] Invalid input tags! must be a valid JSON.\nInput: %s\nError: %s\n", tags, err.Error())
		os.Exit(1)
	}

	keys := utils.SortObjectKeys(tagsMap)

	var mapContent []string
	for _, key := range keys {
		mapContent = append(mapContent, "\""+key+"\"="+"\""+tagsMap[key]+"\"")
	}
	return "{" + strings.Join(mapContent, ",") + "}"
}
