package tagging

import (
	"encoding/json"
	"strings"

	"github.com/env0/terratag/internal/convert"
	"github.com/env0/terratag/internal/tag_keys"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func tagAwsInstance(args TagBlockArgs) (*Result, error) {
	var swappedTagsStrings []string

	tagBlock, err := TagBlock(args)
	if err != nil {
		return nil, err
	}
	swappedTagsStrings = append(swappedTagsStrings, tagBlock)

	volumeTagBlockArgs := args
	volumeTagBlockArgs.TagId = "volume_tags"
	volumeTagBlock, err := TagBlock(volumeTagBlockArgs)
	if err != nil {
		return nil, err
	}
	swappedTagsStrings = append(swappedTagsStrings, volumeTagBlock)

	return &Result{SwappedTagsStrings: swappedTagsStrings}, nil
}

func tagAutoscalingGroup(args TagBlockArgs) (*Result, error) {
	// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html

	var tagsMap map[string]string
	if err := json.Unmarshal([]byte(args.Tags), &tagsMap); err != nil {
		return nil, err
	}

	tagsAttr := args.Block.Body().GetAttribute("tags")
	if tagsAttr != nil {
		// "tags" interpolation is used
		tokens := tagsAttr.Expr().BuildTokens(hclwrite.Tokens{})
		expression := strings.TrimSpace(string(tokens.Bytes()))
		// may be wrapped with ${ } in TF11
		expression = strings.TrimPrefix(expression, "${")
		expression = strings.TrimSuffix(expression, "${")

		key := "local." + tag_keys.GetTerratagAddedKey(args.Filename)
		newTagsValue := "flatten([" + key + "," + expression + "])"

		newTags := ParseHclValueStringToTokens(newTagsValue)

		args.Block.Body().SetAttributeRaw("tags", newTags)
	} else {
		// no "tags" interpolation is used, but rather multiple instances of a "tag" block
		if err := convert.AppendTagBlocks(args.Block, args.Tags); err != nil {
			return nil, err
		}
	}

	return &Result{}, nil
}
