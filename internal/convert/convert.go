package convert

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/env0/terratag/internal/common"
	"github.com/env0/terratag/internal/tag_keys"
	"github.com/env0/terratag/internal/utils"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
	"github.com/zclconf/go-cty/cty"
)

func GetExistingTagsExpression(tokens hclwrite.Tokens) string {
	return stringifyExpression(tokens)
}

func isHclMap(tokens hclwrite.Tokens) bool {
	maybeHclMap := strings.TrimSpace(string(tokens.Bytes()))

	return strings.HasPrefix(maybeHclMap, "{") && strings.HasSuffix(maybeHclMap, "}")
}

func stringifyExpression(tokens hclwrite.Tokens) string {
	expression := strings.TrimSpace(string(tokens.Bytes()))
	// may be wrapped with ${ } in TF11
	expression = strings.TrimPrefix(expression, "${")
	expression = strings.TrimSuffix(expression, "${")

	return expression
}

func AppendLocalsBlock(file *hclwrite.File, filename string, terratag common.TerratagLocal) {
	key := tag_keys.GetTerratagAddedKey(filename)

	// If there's an existings terratag locals replace it with the merged locals.
	blocks := file.Body().Blocks()
	for _, block := range blocks {
		if block.Type() != "locals" {
			continue
		}

		if block.Body().GetAttribute(key) == nil {
			continue
		}

		block.Body().RemoveAttribute(key)
		block.Body().SetAttributeValue(key, cty.StringVal(terratag.Added))

		return
	}

	file.Body().AppendNewline()
	locals := file.Body().AppendNewBlock("locals", nil)
	file.Body().AppendNewline()

	locals.Body().SetAttributeValue(key, cty.StringVal(terratag.Added))
}

func AppendTagBlocks(resource *hclwrite.Block, tags string) error {
	var tagsMap map[string]string
	if err := json.Unmarshal([]byte(tags), &tagsMap); err != nil {
		return err
	}

	keys := utils.SortObjectKeys(tagsMap)

	for _, key := range keys {
		resource.Body().AppendNewline()
		tagBlock := resource.Body().AppendNewBlock("tag", nil)
		tagBlock.Body().SetAttributeValue("key", cty.StringVal(key))
		tagBlock.Body().SetAttributeValue("value", cty.StringVal(tagsMap[key]))
		tagBlock.Body().SetAttributeValue("propagate_at_launch", cty.BoolVal(true))
	}

	return nil
}

func UnquoteTagsAttribute(swappedTagsStrings []string, text string) string {
	for _, swappedTagString := range swappedTagsStrings {
		// treat quotes
		escapedByWriter := strings.ReplaceAll(swappedTagString, "\"", "\\\"")

		// treat variables
		escapedByWriter = strings.ReplaceAll(escapedByWriter, "${", "$${")

		// add quotes if string isn't wrapped with ${}
		if !(strings.HasPrefix(swappedTagString, "${") && strings.HasSuffix(swappedTagString, "}")) {
			escapedByWriter = "\"" + escapedByWriter + "\""
		}

		text = strings.ReplaceAll(text, escapedByWriter, swappedTagString)
	}

	return text
}

func MoveExistingTags(filename string, terratag common.TerratagLocal, block *hclwrite.Block, tagId string) (bool, error) {
	var existingTags hclwrite.Tokens

	// First we try to find tags as attribute
	tagsAttribute := block.Body().GetAttribute(tagId)

	if tagsAttribute != nil {
		// If attribute found, get its value
		log.Print("Pre-existing " + tagId + " ATTRIBUTE found on resource. Merging.")

		existingTags = quoteAttributeKeys(tagsAttribute)
	} else {
		// Otherwise, we try to get tags as block
		tagsBlock := block.Body().FirstMatchingBlock(tagId, nil)
		if tagsBlock != nil {
			quotedTagBlock := quoteBlockKeys(tagsBlock)
			existingTags = funk.Tail(quotedTagBlock.BuildTokens(hclwrite.Tokens{})).(hclwrite.Tokens)
			// If we did get tags from block, we will now remove that block, as we're going to add a merged tags ATTRIBUTE
			removeBlockResult := block.Body().RemoveBlock(tagsBlock)
			if !removeBlockResult {
				return false, errors.New("failed to remove found tags block")
			}
		}
	}

	if existingTags != nil {
		terratag.Found[tag_keys.GetResourceExistingTagsKey(filename, block)] = existingTags

		return true, nil
	}

	return false, nil
}

func quoteBlockKeys(tagsBlock *hclwrite.Block) *hclwrite.Block {
	// In HCL, block keys must NOT be quoted
	// But we need them to be, as we throw them into a map() function as strings
	quotedTagBlock := hclwrite.NewBlock(tagsBlock.Type(), tagsBlock.Labels())
	for key, value := range tagsBlock.Body().Attributes() {
		quotedTagBlock.Body().SetAttributeRaw("\""+key+"\"", value.Expr().BuildTokens(hclwrite.Tokens{}))
	}

	return quotedTagBlock
}

func isTagKeyUnquoted(tags hclwrite.Tokens, index int) bool {
	return tags[index].Type == hclsyntax.TokenIdent && index+1 < len(tags) && tags[index+1].Type == hclsyntax.TokenEqual
}

func quoteAttributeKeys(tagsAttribute *hclwrite.Attribute) hclwrite.Tokens {
	var newTags hclwrite.Tokens

	tags := tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{})

	// if attribute is a variable
	if !(isHclMap(tags)) {
		return tags
	}

	for i, token := range tags {
		if isTagKeyUnquoted(tags, i) {
			openQuote := &hclwrite.Token{
				Type:  hclsyntax.TokenOQuote,
				Bytes: []byte("\""),
				// open quote should have the token ident spaces
				SpacesBefore: token.SpacesBefore,
			}

			closeQuote := &hclwrite.Token{
				Type:         hclsyntax.TokenCQuote,
				Bytes:        []byte("\""),
				SpacesBefore: 0,
			}

			// token ident spaces are now zero since we add an opening quote
			token.SpacesBefore = 0

			newTags = append(newTags, openQuote, token, closeQuote)
		} else {
			newTags = append(newTags, token)
		}
	}

	return newTags
}
