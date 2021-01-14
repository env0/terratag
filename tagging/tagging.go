package tagging

import (
	"github.com/env0/terratag/convert"
	"github.com/env0/terratag/tag_keys"
	"github.com/env0/terratag/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

	mergeCommand := &hclwrite.Token{
		Type:         hclsyntax.TokenIdent,
		Bytes:        []byte("merge"),
		SpacesBefore: 0,
	}

	localPrefix := &hclwrite.Token{
		Type:         hclsyntax.TokenIdent,
		Bytes:        []byte("local"),
		SpacesBefore: 0,
	}

	dotToken := &hclwrite.Token{
		Type:         hclsyntax.TokenDot,
		Bytes:        []byte("."),
		SpacesBefore: 0,
	}

	terraTagLocal := &hclwrite.Token{
		Type:         hclsyntax.TokenIdent,
		Bytes:        []byte(tag_keys.GetTerratagAddedKey(args.Filename)),
		SpacesBefore: 0,
	}

	comma := &hclwrite.Token{
		Type:         hclsyntax.TokenComma,
		Bytes:        []byte(","),
		SpacesBefore: 0,
	}

	openParen := &hclwrite.Token{
		Type:         hclsyntax.TokenOParen,
		Bytes:        []byte("("),
		SpacesBefore: 0,
	} 

	newLine := &hclwrite.Token{
		Type:         hclsyntax.TokenNewline,
		Bytes:        []byte(""),
		SpacesBefore: 0,
	} 
	
	closeParen := &hclwrite.Token{
		Type:         hclsyntax.TokenCParen,
		Bytes:        []byte(")"),
		SpacesBefore: 0,
	} 

	existingTags := args.Block.Body().GetAttribute(args.TagId)
	existingTagsTokens := existingTags.Expr().BuildTokens(hclwrite.Tokens{})

	var newTokens hclwrite.Tokens
	newTokens = append(newTokens, mergeCommand, openParen, newLine)
	newTokens = append(newTokens, existingTagsTokens...)
	newTokens = append(newTokens, newLine, comma, localPrefix, dotToken, terraTagLocal, closeParen)

	for _, token := range existingTagsTokens {
		log.Print(token.Type)
		log.Print(string(token.Bytes))
	}

	// newTagsValueTokens := ParseHclValueStringToTokens(newTagsValue)
	args.Block.Body().SetAttributeRaw(args.TagId, newTokens)

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
