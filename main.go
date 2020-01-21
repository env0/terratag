package main

import (
	"encoding/json"
	"github.com/bmatcuk/doublestar"
	. "github.com/env0/terratag/cli"
	. "github.com/env0/terratag/errors"
	. "github.com/env0/terratag/terraform"
	. "github.com/env0/terratag/tfschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	tags, dir, isMissingArg := InitArgs()

	tfVersion := GetTeraformVersion()

	if isMissingArg {
		return
	}

	tagDirectoryResources(dir, tags, tfVersion)
}

func tagDirectoryResources(dir string, tags string, tfVersion int) {
	matches, err := doublestar.Glob(dir + "/**/*.tf")
	PanicOnError(err, nil)

	for i, match := range matches {
		resolvedMatch, err := filepath.EvalSymlinks(match)
		matches[i] = resolvedMatch
		PanicOnError(err, nil)
	}
	matches = funk.UniqString(matches)

	for _, path := range matches {
		tagFileResources(path, dir, tags, tfVersion)
	}
}

func tagFileResources(path string, dir string, tags string, tfVersion int) {
	log.Print("Processing file ", path)
	hcl := parseHcl(path)

	terratag := TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: jsonToHclMap(tags),
	}

	filename := getFilename(path)

	anyTagged := false
	var swappedTagsStrings []string

	for _, resource := range hcl.Body().Blocks() {
		if resource.Type() == "resource" {
			resourceType := resource.Labels()[0]
			log.Print("Processing resource ", resource.Labels())

			isTaggable, isTaggableViaSpecialTagBlock := IsTaggable(dir, resourceType)

			if isTaggable {
				if !isTaggableViaSpecialTagBlock {
					// for now, we count on it that if there's a single "tag" in the schema (unlike "tags" block),
					// then no "tags" interpolation is used, but rather multiple instances of a "tag" block
					// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html
					swappedTagsStrings = append(swappedTagsStrings, tagResource(filename, terratag, resource, tfVersion))
				} else {
					appendTagBlocks(resource, tags)
				}
				anyTagged = true
			} else {
				log.Print("Resource not taggable, skipping. ")
			}
		}
	}

	if anyTagged {
		appendLocalsBlock(hcl, filename, terratag)

		text := string(hcl.Bytes())

		swappedTagsStrings = append(swappedTagsStrings, terratag.Added)
		text = unquoteTagsAttribute(swappedTagsStrings, text)

		replaceWithTerratagFile(path, text)
	} else {
		log.Print("No taggable resources found in file ", path, " - skipping")
	}
}

func parseHcl(path string) *hclwrite.File {
	src, err := ioutil.ReadFile(path)
	PanicOnError(err, nil)

	file, diagnostics := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if diagnostics.HasErrors() {
		hclErrors := diagnostics.Errs()
		log.Fatalln(hclErrors)
	}
	return file
}

func getFilename(path string) string {
	_, filename := filepath.Split(path)
	filename = strings.TrimSuffix(filename, filepath.Ext(path))
	filename = strings.ReplaceAll(filename, ".", "-")
	return filename
}

func appendLocalsBlock(file *hclwrite.File, filename string, terratag TerratagLocal) {
	file.Body().AppendNewline()
	locals := file.Body().AppendNewBlock("locals", nil)
	file.Body().AppendNewline()

	locals.Body().SetAttributeValue(getTerratagAddedKey(filename), cty.StringVal(terratag.Added))
}

func appendTagBlocks(resource *hclwrite.Block, tags string) {
	var tagsMap map[string]string
	err := json.Unmarshal([]byte(tags), &tagsMap)
	PanicOnError(err, nil)

	for key, value := range tagsMap {
		resource.Body().AppendNewline()
		tagBlock := resource.Body().AppendNewBlock("tag", nil)
		tagBlock.Body().SetAttributeValue("key", cty.StringVal(key))
		tagBlock.Body().SetAttributeValue("value", cty.StringVal(value))
		tagBlock.Body().SetAttributeValue("propagate_at_launch", cty.BoolVal(true))
	}
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

func replaceWithTerratagFile(path string, textContent string) {
	taggedFilename := strings.TrimSuffix(path, filepath.Ext(path)) + ".terratag.tf"
	backupFilename := path + ".bak"

	log.Print("Creating file ", taggedFilename)
	taggedFileError := ioutil.WriteFile(taggedFilename, []byte(textContent), 0644)
	PanicOnError(taggedFileError, nil)

	log.Print("Renaming original file from ", path, " to ", backupFilename)
	backupFileError := os.Rename(path, backupFilename)
	PanicOnError(backupFileError, nil)
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
		tagsValue = "merge( " + getExistingTagsExpression(terratag.Found[getResourceExistingTagsKey(filename, resource)]) + ", local." + getTerratagAddedKey(filename) + ")"
	} else {
		tagsValue = "local." + getTerratagAddedKey(filename)
	}

	if tfVersion == 11 {
		tagsValue = "${" + tagsValue + "}"
	}

	resource.Body().SetAttributeValue("tags", cty.StringVal(tagsValue))

	return tagsValue
}

func getExistingTagsExpression(tokens hclwrite.Tokens) string {
	if isHclMap(tokens) {
		return buildMapExpression(tokens)
	} else {
		return stringifyExpression(tokens)
	}
}

func isHclMap(tokens hclwrite.Tokens) bool {
	maybeHclMap := strings.TrimSpace(string(tokens.Bytes()))
	return strings.HasPrefix(maybeHclMap, "{") && strings.HasSuffix(maybeHclMap, "}")
}

func buildMapExpression(tokens hclwrite.Tokens) string {
	// Need to convert to inline map expression

	// First, we remove the first and last two tokens - get rid of the openning { closing } and newline
	tokens = tokens[1:]
	tokens = tokens[:len(tokens)-2]

	// Then, we normalize the key-value paris so that they would be seperated by comma
	// That's cause HCL supports both newline, comma or a combination of the two in seperating values
	// This will make it easier to split the key-value pairs later
	var tokensToRemove hclwrite.Tokens
	for i, token := range tokens {
		if token.Type == hclsyntax.TokenNewline {
			// make sure there is (or was) no comma before
			if i > 0 && tokens[i-1].Type != hclsyntax.TokenComma && !funk.Contains(tokensToRemove, tokens[i-1]) {
				tokens[i] = &hclwrite.Token{
					Type:         hclsyntax.TokenComma,
					Bytes:        []byte(","),
					SpacesBefore: 1,
				}
			} else { // if there is, we should remove this new line, so we'll only have the comma
				tokensToRemove = append(tokensToRemove, token)
			}
		}
	}

	// Remove tall the new lines we marked for removal
	for _, tokenToRemove := range tokensToRemove {
		indexToRemove := funk.IndexOf(tokens, tokenToRemove)
		tokens = append(tokens[:indexToRemove], tokens[indexToRemove+1:]...)
	}

	// At this point there should be no new lines, only a single comma seperating between key values
	// Since the map() gets a flat set of pairs as map("key1","value1","key2","value2"),
	// we can just replace any assignment operator (=) with comma
	for i, token := range tokens {
		if token.Type == hclsyntax.TokenEqual {
			tokens[i] = &hclwrite.Token{
				Type:         hclsyntax.TokenComma,
				Bytes:        []byte(","),
				SpacesBefore: token.SpacesBefore,
			}
		}
	}

	mapContent := string(tokens.Bytes())
	mapContent = strings.TrimSpace(mapContent)
	mapContent = strings.TrimSuffix(mapContent, ",") // remove any traling commas due to newline replaced
	return "map(" + mapContent + ")"
}

func stringifyExpression(tokens hclwrite.Tokens) string {
	expression := strings.TrimSpace(string(tokens.Bytes()))
	// may be wrapped with ${ } in TF11
	expression = strings.TrimPrefix(expression, "${")
	expression = strings.TrimSuffix(expression, "${")

	return expression
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
			quaotedTagBlock := quoteBlockKeys(tagsBlock)
			existingTags = funk.Tail(quaotedTagBlock.BuildTokens(hclwrite.Tokens{})).(hclwrite.Tokens)
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

func quoteBlockKeys(tagsBlock *hclwrite.Block) *hclwrite.Block {
	// In HCL, block keys must NOT be quoted
	// But we need them to be, as we throw them into a map() function as strings
	quaotedTagBlock := hclwrite.NewBlock(tagsBlock.Type(), tagsBlock.Labels())
	for key, value := range tagsBlock.Body().Attributes() {
		quaotedTagBlock.Body().SetAttributeRaw("\""+key+"\"", value.Expr().BuildTokens(hclwrite.Tokens{}))
	}
	return quaotedTagBlock
}

func getTerratagAddedKey(filname string) string {
	return "terratag_added_" + filname
}

func getResourceExistingTagsKey(filename string, resource *hclwrite.Block) string {
	delimiter := "__"
	return "terratag_found_" + filename + delimiter + strings.Join(resource.Labels(), delimiter)
}

type TerratagLocal struct {
	Found map[string]hclwrite.Tokens
	Added string
}
