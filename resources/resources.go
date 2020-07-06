package resources

import (
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/tagging"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"log"
)

func HasResourceTagFunc(resourceType string) bool {
	return getResourceTagFunc(resourceType) != nil
}

func getResourceTagFunc(resourceType string) TagResourceFunc {
	return resourceMap[resourceType]
}

func TagBlockAndGetTagStrings(args tagging.TagBlockArgs) []string {
	var swappedTagsStrings []string
	resourceType := terraform.GetResourceType(*args.Block)

	customResourceTaggingFunc := getResourceTagFunc(resourceType)

	if customResourceTaggingFunc != nil {
		swappedTagsStrings = customResourceTaggingFunc(args)
	} else {
		swappedTagsStrings = append(swappedTagsStrings, tagging.TagBlock(args))
	}

	return swappedTagsStrings
}

func tagGcpContainerNodePool(args tagging.TagBlockArgs) []string {
	var blocks []*hclwrite.Block
	var swappedTagsStrings []string

	nodeConfig := terraform.GetSingleNestedBlock(args.Block, "node_config")
	if nodeConfig != nil {
		log.Print("Found taggable nested node_config block, processing...")
		blocks = append(blocks, nodeConfig)
	}

	for _, block := range blocks {
		args.Block = block
		swappedTagsStrings = append(swappedTagsStrings, tagging.TagBlock(args))
	}

	return swappedTagsStrings
}

func tagGcpContainerCluster(args tagging.TagBlockArgs) []string {
	var blocks []*hclwrite.Block
	var swappedTagsStrings []string

	// handle root block
	rootBlockArgs := args
	rootBlockArgs.TagId = "resource_labels"
	swappedTagsStrings = append(swappedTagsStrings, tagging.TagBlock(rootBlockArgs))

	nodeConfig := terraform.GetSingleNestedBlock(args.Block, "node_config")
	if nodeConfig != nil {
		log.Print("Found taggable nested node_config block, processing...")
		blocks = append(blocks, nodeConfig)
	}

	nodeConfig = terraform.GetNestedBlock(args.Block, []string{"node_pool", "node_config"})
	if nodeConfig != nil {
		log.Print("Found taggable nested node_pool/node_config block, processing...")
		blocks = append(blocks, nodeConfig)
	}

	for _, block := range blocks {
		args.Block = block
		swappedTagsStrings = append(swappedTagsStrings, tagging.TagBlock(args))
	}

	return swappedTagsStrings
}

func tagAwsAutoscalingGroup(args tagging.TagBlockArgs) []string {
	var emptyArray []string
	// for now, we count on it that if there's a single "tag" in the schema (unlike "tags" block),
	// then no "tags" interpolation is used, but rather multiple instances of a "tag" block
	// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html
	convert.AppendTagBlocks(args.Block, args.Tags)

	return emptyArray
}

var resourceMap = map[string]TagResourceFunc{
	"aws_autoscaling_group":      tagAwsAutoscalingGroup,
	"google_container_cluster":   tagGcpContainerCluster,
	"google_container_node_pool": tagGcpContainerNodePool,
}

type TagResourceFunc func(args tagging.TagBlockArgs) []string
