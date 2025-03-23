package tagging

import (
	"testing"

	"github.com/env0/terratag/internal/common"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
)

func TestTagBlock_MergeOrder(t *testing.T) {
	testCases := []struct {
		name             string
		keepExistingTags bool
		expectedMerge    string // The complete expected merge expression
	}{
		{
			name:             "Default behavior - new tags override existing",
			keepExistingTags: false,
			// When keepExistingTags=false, existing tags should come first in merge
			expectedMerge: "merge( { \"Name\" = \"Original Name\", \"Environment\" = \"Dev\" }, local.terratag_added_main)",
		},
		{
			name:             "Keep existing tags - existing tags override new",
			keepExistingTags: true,
			// When keepExistingTags=true, new tags should come first in merge
			expectedMerge: "merge( local.terratag_added_main, { \"Name\" = \"Original Name\", \"Environment\" = \"Dev\" })",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh block for each test case
			f := hclwrite.NewEmptyFile()
			rootBody := f.Body()
			resourceBlock := rootBody.AppendNewBlock("resource", []string{"aws_s3_bucket", "test"})

			// Add original tags
			tagsTokens := ParseHclValueStringToTokens(`{ Name = "Original Name", Environment = "Dev" }`)
			resourceBlock.Body().SetAttributeRaw("tags", tagsTokens)

			// Set up terratag local with extracted existing tags
			existingTagsToken := ParseHclValueStringToTokens(`{
				"Name" = "Original Name",
				"Environment" = "Dev"
			}`)

			terratag := common.TerratagLocal{
				Found: map[string]hclwrite.Tokens{
					"terratag_existing_tags_main_resource_aws_s3_bucket_test": existingTagsToken,
				},
				Added: `{"Name"="Terratag Name","Owner"="DevOps"}`,
			}

			args := TagBlockArgs{
				Filename:         "main",
				Block:            resourceBlock,
				Tags:             `{"Name": "Terratag Name", "Owner": "DevOps"}`,
				Terratag:         terratag,
				TagId:            "tags",
				KeepExistingTags: tc.keepExistingTags,
			}

			result, err := TagBlock(args)
			assert.NoError(t, err)

			// Compare the exact merge string
			assert.Equal(t, tc.expectedMerge, result, "The merge expression doesn't match expected value")
		})
	}
}
