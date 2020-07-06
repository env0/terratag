package tagging

import (
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/tag_keys"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TagBlock(args TagBlockArgs) string {
	hasExistingTags := convert.MoveExistingTags(args.Filename, args.Terratag, args.Block, args.TagId)

	tagsValue := ""
	if hasExistingTags {
		tagsValue = "merge( " + convert.GetExistingTagsExpression(args.Terratag.Found[tag_keys.GetResourceExistingTagsKey(args.Filename, args.Block)]) + ", local." + tag_keys.GetTerratagAddedKey(args.Filename) + ")"
	} else {
		tagsValue = "local." + tag_keys.GetTerratagAddedKey(args.Filename)
	}

	if args.TfVersion == 11 {
		tagsValue = "${" + tagsValue + "}"
	}

	args.Block.Body().SetAttributeValue(args.TagId, cty.StringVal(tagsValue))

	return tagsValue
}

type TagBlockArgs struct {
	Filename  string
	Block     *hclwrite.Block
	Tags      string
	Terratag  convert.TerratagLocal
	TagId     string
	TfVersion int
}
