package tagging

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func getNodeConfigBlocks(parent *hclwrite.Block) []*hclwrite.Block {
	var blocks []*hclwrite.Block

	for _, block := range parent.Body().Blocks() {
		if block.Type() == "node_config" {
			blocks = append(blocks, block)
		} else {
			blocks = append(blocks, getNodeConfigBlocks(block)...)
		}
	}

	return blocks
}

func tagNodeConfigBlocks(args TagBlockArgs) []string {
	var swappedTagsStrings []string

	for _, block := range getNodeConfigBlocks(args.Block) {
		args.Block = block
		swappedTagsStrings = append(swappedTagsStrings, TagBlock(args))
	}

	return swappedTagsStrings
}

func tagContainerCluster(args TagBlockArgs) Result {
	var swappedTagsStrings []string

	// handle root block
	rootBlockArgs := args
	rootBlockArgs.TagId = "resource_labels"
	swappedTagsStrings = append(swappedTagsStrings, TagBlock(rootBlockArgs))

	swappedTagsStrings = append(swappedTagsStrings, tagNodeConfigBlocks(args)...)

	return Result{SwappedTagsStrings: swappedTagsStrings}
}

func tagContainerNodePool(args TagBlockArgs) Result {
	return Result{SwappedTagsStrings: tagNodeConfigBlocks(args)}
}
