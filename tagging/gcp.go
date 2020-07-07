package tagging

import (
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"log"
)

func tagContainerCluster(args TagBlockArgs) Result {
	var blocks []*hclwrite.Block
	var swappedTagsStrings []string

	// handle root block
	rootBlockArgs := args
	rootBlockArgs.TagId = "resource_labels"
	swappedTagsStrings = append(swappedTagsStrings, TagBlock(rootBlockArgs))

	nodeConfig := terraform.GetNestedBlock(args.Block, "node_config")
	if nodeConfig != nil {
		log.Print("Found taggable nested node_config block, processing...")
		blocks = append(blocks, nodeConfig)
	}

	nodeConfig = terraform.GetNestedBlock(args.Block, "node_pool", "node_config")
	if nodeConfig != nil {
		log.Print("Found taggable nested node_pool/node_config block, processing...")
		blocks = append(blocks, nodeConfig)
	}

	for _, block := range blocks {
		args.Block = block
		swappedTagsStrings = append(swappedTagsStrings, TagBlock(args))
	}

	return Result{SwappedTagsStrings: swappedTagsStrings}
}

func tagContainerNodePool(args TagBlockArgs) Result {
	var blocks []*hclwrite.Block
	var swappedTagsStrings []string

	nodeConfig := terraform.GetNestedBlock(args.Block, "node_config")
	if nodeConfig != nil {
		log.Print("Found taggable nested node_config block, processing...")
		blocks = append(blocks, nodeConfig)
	}

	for _, block := range blocks {
		args.Block = block
		swappedTagsStrings = append(swappedTagsStrings, TagBlock(args))
	}

	return Result{SwappedTagsStrings: swappedTagsStrings}
}
