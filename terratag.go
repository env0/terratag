package terratag

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/env0/terratag/cli"
	"github.com/env0/terratag/internal/common"
	"github.com/env0/terratag/internal/convert"
	"github.com/env0/terratag/internal/file"
	"github.com/env0/terratag/internal/providers"
	"github.com/env0/terratag/internal/tag_keys"
	"github.com/env0/terratag/internal/tagging"
	"github.com/env0/terratag/internal/terraform"
	"github.com/env0/terratag/internal/tfschema"
	"github.com/env0/terratag/internal/utils"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type counters struct {
	totalResources  uint32
	taggedResources uint32
	totalFiles      uint32
	taggedFiles     uint32
}

var pairRegex = regexp.MustCompile(`^([a-zA-Z][\w-]*)=([\w-]+)$`)

var matchWaitGroup sync.WaitGroup

func (c *counters) Add(other counters) {
	atomic.AddUint32(&c.totalResources, other.totalResources)
	atomic.AddUint32(&c.taggedResources, other.taggedResources)
	atomic.AddUint32(&c.totalFiles, other.totalFiles)
	atomic.AddUint32(&c.taggedFiles, other.taggedFiles)
}

func Terratag(args cli.Args) error {
	tfVersion, err := terraform.GetTerraformVersion()
	if err != nil {
		return err
	}

	if err := terraform.ValidateInitRun(args.Dir, args.Type); err != nil {
		return err
	}

	matches, err := terraform.GetFilePaths(args.Dir, args.Type)
	if err != nil {
		return err
	}

	taggingArgs := &common.TaggingArgs{
		Filter:              args.Filter,
		InvertFilter:        args.InvertFilter,
		Dir:                 args.Dir,
		Tags:                args.Tags,
		Matches:             matches,
		IsSkipTerratagFiles: args.IsSkipTerratagFiles,
		Rename:              args.Rename,
		IACType:             common.IACType(args.Type),
		TFVersion:           *tfVersion,
	}

	counters := tagDirectoryResources(taggingArgs)
	log.Print("[INFO] Summary:")
	log.Print("[INFO] Tagged ", counters.taggedResources, " resource/s (out of ", counters.totalResources, " resource/s processed)")
	log.Print("[INFO] In ", counters.taggedFiles, " file/s (out of ", counters.totalFiles, " file/s processed)")

	return nil
}

// dir string, filter string, matches []string, tags string, isSkipTerratagFiles bool, tfVersion convert.Version, rename bool, iacType string
func tagDirectoryResources(args *common.TaggingArgs) counters {
	var total counters
	for _, path := range args.Matches {
		if args.IsSkipTerratagFiles && strings.HasSuffix(path, "terratag.tf") {
			log.Print("[INFO] Skipping file ", path, " as it's already tagged")
		} else {
			matchWaitGroup.Add(1)

			go func(path string) {
				defer matchWaitGroup.Done()

				total.Add(counters{
					totalFiles: 1,
				})

				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] failed to process %s due to an exception\n%v", path, r)
					}
				}()

				perFile, err := tagFileResources(path, args)
				if err != nil {
					log.Printf("[ERROR] failed to process %s due to an error\n%v", path, err)
					return
				}

				total.Add(*perFile)
			}(path)
		}
	}

	matchWaitGroup.Wait()

	return total
}

func tagFileResources(path string, args *common.TaggingArgs) (*counters, error) {
	perFileCounters := counters{}
	log.Print("[INFO] Processing file ", path)
	var swappedTagsStrings []string

	hcl, err := file.ReadHCLFile(path)
	if err != nil {
		return nil, err
	}

	filename := file.GetFilename(path)

	hclMap, err := toHclMap(args.Tags)
	if err != nil {
		return nil, err
	}

	terratag := common.TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: hclMap,
	}

	for _, resource := range hcl.Body().Blocks() {
		switch resource.Type() {
		case "resource":
			log.Print("[INFO] Processing resource ", resource.Labels())
			perFileCounters.totalResources += 1

			matched, err := regexp.MatchString(args.Filter, resource.Labels()[0])
			if err != nil {
				return nil, err
			}

			// invert the match if the invert Filter flag is set
			if args.InvertFilter {
				matched = !matched
			}

			if !matched {
				log.Print("[INFO] Resource excluded by filter, skipping.", resource.Labels())
				continue
			}

			if args.InvertFilter {
				matched = !matched
			}

			isTaggable, err := tfschema.IsTaggable(args.Dir, args.IACType, *resource)
			if err != nil {
				return nil, err
			}

			if isTaggable {
				log.Print("[INFO] Resource taggable, processing...", resource.Labels())
				perFileCounters.taggedResources += 1
				result, err := tagging.TagResource(tagging.TagBlockArgs{
					Filename:  filename,
					Block:     resource,
					Tags:      args.Tags,
					Terratag:  terratag,
					TagId:     providers.GetTagIdByResource(terraform.GetResourceType(*resource)),
					TfVersion: args.TFVersion,
				})
				if err != nil {
					return nil, err
				}

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
						return nil, err
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

		if err := file.ReplaceWithTerratagFile(path, text, args.Rename); err != nil {
			return nil, err
		}
		perFileCounters.taggedFiles = 1
	} else {
		log.Print("[INFO] No taggable resources found in file ", path, " - skipping")
	}
	return &perFileCounters, nil
}

func toHclMap(tags string) (string, error) {
	var tagsMap map[string]string
	err := json.Unmarshal([]byte(tags), &tagsMap)
	if err != nil {
		// If it's not a JSON it might be "key1=value1,key2=value2".
		tagsMap = make(map[string]string)
		pairs := strings.Split(tags, ",")
		for _, pair := range pairs {
			match := pairRegex.FindStringSubmatch(pair)
			if match == nil {
				return "", fmt.Errorf("invalid input tags! must be a valid JSON or pairs of key=value.\nInput: %s", tags)
			}
			tagsMap[match[1]] = match[2]
		}
	}

	keys := utils.SortObjectKeys(tagsMap)

	var mapContent []string
	for _, key := range keys {
		mapContent = append(mapContent, "\""+key+"\"="+"\""+tagsMap[key]+"\"")
	}
	return "{" + strings.Join(mapContent, ",") + "}", nil
}
