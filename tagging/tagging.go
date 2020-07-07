package tagging

import (
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/tag_keys"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func DefaultTaggingFn(args TagBlockArgs) Result {
	return Result{
		SwappedTagsStrings: []string{TagBlock(args)},
	}
}

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

func HasResourceTagFn(resourceType string) bool {
	return resourceTypeToFnMap[resourceType] != nil
}

func GetTaggingResult(args TagBlockArgs) Result {
	var result Result
	resourceType := terraform.GetResourceType(*args.Block)

	customTaggingFn := resourceTypeToFnMap[resourceType]

	if customTaggingFn != nil {
		result = customTaggingFn(args)
	} else {
		result = DefaultTaggingFn(args)
	}

	return result
}

var resourceTypeToFnMap = map[string]TagResourceFn{
	"aws_autoscaling_group":      tagAutoscalingGroup,
	"google_container_cluster":   tagContainerCluster,
	"google_container_node_pool": tagContainerNodePool,
}

type TagResourceFn func(args TagBlockArgs) Result

type Result struct {
	SwappedTagsStrings []string
}
