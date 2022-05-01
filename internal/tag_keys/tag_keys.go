package tag_keys

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"strings"
)

func GetTerratagAddedKey(filname string) string {
	return "terratag_added_" + filname
}

func GetResourceExistingTagsKey(filename string, resource *hclwrite.Block) string {
	delimiter := "__"
	return "terratag_found_" + filename + delimiter + strings.Join(resource.Labels(), delimiter)
}
