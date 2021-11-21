package tagging

import (
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/tag_keys"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"log"
)

func defaultTaggingFn(args TagBlockArgs) Result {
	return Result{
		SwappedTagsStrings: []string{TagBlock(args)},
	}
}

func ParseHclValueStringToTokens(hclValueString string) hclwrite.Tokens {
	file, diags := hclwrite.ParseConfig([]byte("tempKey = "+hclValueString), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		log.Print("error parsing hcl value string " + hclValueString)
		panic(diags.Errs()[0])
	}
	tempAttribute := file.Body().GetAttribute("tempKey")
	return tempAttribute.Expr().BuildTokens(hclwrite.Tokens{})
}

func TagBlock(args TagBlockArgs) string {
	hasExistingTags := convert.MoveExistingTags(args.Filename, args.Terratag, args.Block, args.TagId)

	terratagAddedKey := "local." + tag_keys.GetTerratagAddedKey(args.Filename)
	newTagsValue := terratagAddedKey

	if hasExistingTags {
		existingTagsExpression := getExistingTagsExpression(args)
		newTagsValue = "merge( " + existingTagsExpression + ", " + terratagAddedKey + ")"
	}

	newTagsValue = fixLegacyTerraform(args, newTagsValue)

	newTagsValueTokens := ParseHclValueStringToTokens(newTagsValue)
	args.Block.Body().SetAttributeRaw(args.TagId, newTagsValueTokens)

	return newTagsValue
}

func ConcatTagToTagsBlock(args TagBlockArgs) string {
	convert.MoveExistingTags(args.Filename, args.Terratag, args.Block, args.TagId)

	terratagAddedKey := "local." + tag_keys.GetTerratagAddedKey(args.Filename)
	newTagsValue := terratagAddedKey

	existingTagsExpression := getExistingTagsExpression(args)
	newTagsValue = "concat( " + existingTagsExpression + ", [" + terratagAddedKey + "])"

	newTagsValue = fixLegacyTerraform(args, newTagsValue)

	newTagsValueTokens := ParseHclValueStringToTokens(newTagsValue)
	args.Block.Body().SetAttributeRaw(args.TagId, newTagsValueTokens)

	return newTagsValue
}

func HasResourceTagFn(resourceType string) bool {
	return resourceTypeToFnMap[resourceType] != nil
}

func TagResource(args TagBlockArgs) Result {
	var result Result
	resourceType := terraform.GetResourceType(*args.Block)

	customTaggingFn := resourceTypeToFnMap[resourceType]

	if customTaggingFn != nil {
		result = customTaggingFn(args)
	} else {
		result = defaultTaggingFn(args)
	}

	return result
}

func getExistingTagsExpression(args TagBlockArgs) string {
	existingTagsKey := tag_keys.GetResourceExistingTagsKey(args.Filename, args.Block)
	existingTagsExpression := convert.GetExistingTagsExpression(args.Terratag.Found[existingTagsKey], args.TfVersion)
	return existingTagsExpression
}

func fixLegacyTerraform(args TagBlockArgs, newTagsValue string) string {
	if args.TfVersion.Major == 0 && args.TfVersion.Minor == 11 {
		newTagsValue = "\"${" + newTagsValue + "}\""
	}
	return newTagsValue
}

var resourceTypeToFnMap = map[string]TagResourceFn{
	"aws_autoscaling_group":      tagAutoscalingGroup,
	"google_container_cluster":   tagContainerCluster,
	"azurerm_kubernetes_cluster": tagAksK8sCluster,
}

type TagBlockArgs struct {
	Filename  string
	Block     *hclwrite.Block
	Tags      string
	Terratag  convert.TerratagLocal
	TagId     string
	TfVersion convert.Version
}

type TagResourceFn func(args TagBlockArgs) Result

type Result struct {
	SwappedTagsStrings []string
}
