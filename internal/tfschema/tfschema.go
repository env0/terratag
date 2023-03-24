package tfschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/env0/terratag/internal/common"
	"github.com/env0/terratag/internal/providers"
	"github.com/env0/terratag/internal/tagging"
	"github.com/env0/terratag/internal/terraform"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

var ErrResourceTypeNotFound = errors.New("resource type not found")

var providerSchemasMap map[string]*ProviderSchemas = map[string]*ProviderSchemas{}
var providerSchemasMapLock sync.Mutex

//var customSupportedProviderNames = [...]string{"google-beta"}

type Attribute struct {
	Type      cty.Type `json:"type"`
	Required  bool     `json:"required"`
	Optional  bool     `json:"optional"`
	Computed  bool     `json:"computed"`
	Sensitive bool     `json:"sensitive"`
}

type Block struct {
	Attributes map[string]*Attribute `json:"attributes"`
}

type ResourceSchema struct {
	Block Block `json:"block"`
}

type ProviderSchema struct {
	ResourceSchemas map[string]*ResourceSchema `json:"resource_schemas"`
}

type ProviderSchemas struct {
	ProviderSchemas map[string]*ProviderSchema `json:"provider_schemas"`
}

func IsTaggable(dir string, iacType common.IACType, resource hclwrite.Block) (bool, error) {
	var isTaggable bool
	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType) {
		resourceSchema, err := getResourceSchema(resourceType, dir, iacType)
		if err != nil {
			if err == ErrResourceTypeNotFound {
				log.Print("[WARN] Skipped ", resourceType, " as it is not YET supported")
				return false, nil
			}

			return false, err
		}

		for attribute := range resourceSchema.Block.Attributes {
			if providers.IsTaggableByAttribute(resourceType, attribute) {
				isTaggable = true
			}
		}

		if tagging.HasResourceTagFn(resourceType) {
			isTaggable = true
		}
	}

	return isTaggable, nil
}

type TfSchemaAttribute struct {
	Name string
	Type string
}

func getTerragruntPluginPath(dir string) string {
	dir += "/.terragrunt-cache"
	ret := dir
	found := false

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if found || err != nil {
			return filepath.SkipDir
		}

		// E.g. ./.terragrunt-cache/yHtqnMrVQOISIYxobafVvZbAAyU/ThyYwttwki6d6AS3aD5OwoyqIWA/.terraform
		if strings.HasSuffix(path, "/.terraform") {
			ret = strings.TrimSuffix(path, "/.terraform")
			found = true
		}

		return nil
	})

	return ret
}

func getResourceSchema(resourceType string, dir string, iacType common.IACType) (*ResourceSchema, error) {
	if iacType == common.Terragrunt {
		// which mode of terragrunt it is (with or without cache folder).
		if _, err := os.Stat(dir + "/.terragrunt-cache"); err == nil {
			dir = getTerragruntPluginPath(dir)
		}
	}

	providerSchemasMapLock.Lock()
	defer providerSchemasMapLock.Unlock()

	providerSchemas, ok := providerSchemasMap[dir]
	if !ok {
		providerSchemas = &ProviderSchemas{}

		cmd := exec.Command("terraform", "providers", "schema", "-json")
		cmd.Dir = dir

		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to execute 'terraform providers schema -json' command: %w", err)
		}

		// Remove any command output "junk" in the prefix.
		if start := bytes.Index(out, []byte("{")); start != -1 {
			out = out[start:]
		}

		// Remove any command output "junk" in the suffix.
		if end := bytes.LastIndex(out, []byte("}")); end != -1 {
			out = out[0 : end+1]
		}

		if err := json.Unmarshal(out, providerSchemas); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d %s", e.Offset, string(out)[e.Offset-100:e.Offset+1])
			}
			return nil, fmt.Errorf("failed to unmarshal returned provider schemas: %w", err)
		}

		providerSchemasMap[dir] = providerSchemas
	}

	for _, providerSchema := range providerSchemas.ProviderSchemas {
		resourceSchema, ok := providerSchema.ResourceSchemas[resourceType]
		if ok {
			return resourceSchema, nil
		}
	}

	return nil, ErrResourceTypeNotFound
}
