package tagging

import (
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/tag_keys"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcl/v2"
	"log"
)

func defaultTaggingFn(args TagBlockArgs) Result {
	return Result{
		SwappedTagsStrings: []string{TagBlock(args)},
	}
}

func ParseHclValueStringToTokens(hclValueString string) hclwrite.Tokens {
	file, diags := hclwrite.ParseConfig([]byte("tempKey = " + hclValueString), "", hcl.Pos{Line: 1, Column: 1})
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
		existingTagsKey := tag_keys.GetResourceExistingTagsKey(args.Filename, args.Block)
		existingTagsExpression := convert.GetExistingTagsExpression(args.Terratag.Found[existingTagsKey])
		newTagsValue = "merge( " + existingTagsExpression + ", " + terratagAddedKey + ")"
	}

	if args.TfVersion == 11 {
		newTagsValue = "${" + newTagsValue + "}"
	}

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
	TfVersion int
}

type TagResourceFn func(args TagBlockArgs) Result

type Result struct {
	SwappedTagsStrings []string
}
