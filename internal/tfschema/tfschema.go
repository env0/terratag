package tfschema

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/env0/terratag/internal/common"
	"github.com/env0/terratag/internal/providers"
	"github.com/env0/terratag/internal/tagging"
	"github.com/env0/terratag/internal/terraform"
	"github.com/thoas/go-funk"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfschema/tfschema"
)

var providerToClientMap = map[string]tfschema.Client{}
var providerToClientMapLock sync.Mutex

var customSupportedProviderNames = [...]string{"google-beta"}

func IsTaggable(dir string, iacType common.IACType, resource hclwrite.Block) (bool, error) {
	var isTaggable bool
	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType) {
		providerName, _ := detectProviderName(resource)
		client, err := getClient(providerName, dir, iacType)
		if err != nil {
			return false, err
		}
		typeSchema, err := client.GetResourceTypeSchema(resourceType)
		if err != nil {
			if strings.Contains(err.Error(), "Failed to find resource type") {
				// short circuiting unfound resource due to: https://github.com/env0/terratag/issues/17
				log.Print("[WARN] Skipped ", resourceType, " as it is not YET supported")
				return false, nil
			}

			return false, err
		}

		attributes := typeSchema.Attributes
		for attribute := range attributes {
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

func extractProviderNameFromResourceType(resourceType string) (string, error) {
	s := strings.SplitN(resourceType, "_", 2)
	if len(s) < 2 {
		return "", fmt.Errorf("failed to detect a provider name: %s", resourceType)
	}
	return s[0], nil
}

func detectProviderName(resource hclwrite.Block) (string, error) {
	providerAttribute := resource.Body().GetAttribute("provider")

	if providerAttribute != nil {
		providerTokens := providerAttribute.Expr().BuildTokens(hclwrite.Tokens{})
		providerName := strings.Trim(string(providerTokens.Bytes()), "\" ")
		if funk.Contains(customSupportedProviderNames, providerName) {
			return providerName, nil
		}
	}

	return extractProviderNameFromResourceType(terraform.GetResourceType(resource))
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

func getClient(providerName string, dir string, iacType common.IACType) (tfschema.Client, error) {
	if iacType == common.Terragrunt {
		// Check which mode of terragrunt it is (with or without cache folder).
		if _, err := os.Stat("/.terragrunt-cache"); err == nil {
			dir = getTerragruntPluginPath(dir)
		}
	}

	providerToClientMapLock.Lock()
	defer providerToClientMapLock.Unlock()

	client, exists := providerToClientMap[providerName]
	if exists {
		return client, nil
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		// this annoyance - both tfschema and go-plugin open output
		// directly to os.Stderr, bypassing our log level filter.
		// weird to need to bypass the issue by assigning the default
		// output ¯\_(ツ)_/¯
	})

	newClient, err := tfschema.NewClient(providerName, tfschema.Option{
		RootDir: dir,
		Logger:  logger,
	})
	if err != nil {
		return nil, err
	}

	providerToClientMap[providerName] = newClient

	return newClient, nil
}
