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

	// Tag 'volume_tags' if it exists.
	// Else:
	//  1. create 'root_block_device' block (if not exist) and add tags to it.
	//  2. add tags to any existing 'ebs_block_device' block.
	// See tag guide for additional details: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance#tag-guide

	if args.Block.Body().GetAttribute("volume_tags") != nil {
		// Add tags to 'volume_tags' attribute.
		volumeTagBlockArgs := args
		volumeTagBlockArgs.TagId = "volume_tags"
		volumeTagBlock, err := TagBlock(volumeTagBlockArgs)
		if err != nil {
			return nil, err
		}
		swappedTagsStrings = append(swappedTagsStrings, volumeTagBlock)
	} else {
		rootBlockDevice := args.Block.Body().FirstMatchingBlock("root_block_device", nil)
		if rootBlockDevice == nil {
			// Create 'root_block_device' block.
			rootBlockDevice = args.Block.Body().AppendNewBlock("root_block_device", nil)
		}

		// Add tags to 'root_block_device' block.
		origArgsBlock := args.Block
		args.Block = rootBlockDevice
		tagBlock, err := TagBlock(args)
		if err != nil {
			return nil, err
		}
		swappedTagsStrings = append(swappedTagsStrings, tagBlock)
		args.Block = origArgsBlock

		// Add tags to any 'ebs_block_device' blocks (if any exist).
		for _, block := range args.Block.Body().Blocks() {
			if block.Type() != "ebs_block_device" {
				continue
			}

			origArgsBlock := args.Block
			args.Block = block
			tagBlock, err := TagBlock(args)
			if err != nil {
				return nil, err
			}
			swappedTagsStrings = append(swappedTagsStrings, tagBlock)
			args.Block = origArgsBlock
		}
	}

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
