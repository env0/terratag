package terraform

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/env0/terratag/internal/common"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
)

func GetResourceType(resource hclwrite.Block) string {
	return resource.Labels()[0]
}

func getRootDir(dir string, iacType string) string {
	if iacType == string(common.Terragrunt) {
		// which mode of terragrunt it is (with or without cache folder).
		if _, err := os.Stat(dir + "/.terragrunt-cache"); err != nil {
			if os.IsNotExist(err) {
				return ""
			}
		}

		return "/.terragrunt-cache"
	} else {
		return "/.terraform"
	}
}

func ValidateInitRun(dir string, iacType string) error {
	path := dir + getRootDir(dir, iacType)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s init must run before running terratag", iacType)
		}

		return fmt.Errorf("couldn't determine if %s init has run: %w", iacType, err)
	}

	return nil
}

func GetFilePaths(dir string, iacType string) ([]string, error) {
	if iacType == string(common.Terragrunt) {
		return getTerragruntFilePath(dir)
	} else {
		return getTerraformFilePaths(dir)
	}
}

func getTerragruntFilePath(rootDir string) ([]string, error) {
	var tfFiles []string

	if err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("[WARN] skipping %s due to an error: %v", path, err)

			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".tf") {
			tfFiles = append(tfFiles, path)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return tfFiles, nil
}

func getTerraformFilePaths(rootDir string) ([]string, error) {
	const tfFileMatcher = "/*.tf"

	tfFiles, err := doublestar.Glob(rootDir + tfFileMatcher)
	if err != nil {
		return nil, err
	}

	modulesDirs, err := getTerraformModulesDirPaths(rootDir)
	if err != nil {
		return nil, err
	}

	for _, moduleDir := range modulesDirs {
		matches, err := doublestar.Glob(moduleDir + tfFileMatcher)
		if err != nil {
			return nil, err
		}

		tfFiles = append(tfFiles, matches...)
	}

	for i, tfFile := range tfFiles {
		resolvedTfFile, err := filepath.EvalSymlinks(tfFile)
		if err != nil {
			return nil, err
		}

		tfFiles[i] = resolvedTfFile
	}

	return funk.UniqString(tfFiles), nil
}

func getTerraformModulesDirPaths(dir string) ([]string, error) {
	paths := []string{}
	modulesJson := ModulesJson{}

	jsonFile, err := os.Open(dir + "/.terraform/modules/modules.json")

	//lint:ignore SA5001 not required to check file close status.
	defer jsonFile.Close()

	if os.IsNotExist(err) {
		return paths, nil
	}

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(byteValue, &modulesJson); err != nil {
		return nil, err
	}

	for _, module := range modulesJson.Modules {
		modulePath, err := filepath.EvalSymlinks(dir + "/" + module.Dir)
		if os.IsNotExist(err) {
			log.Print("[WARN] Module not found, skipping.", dir+"/"+module.Dir)

			continue
		}

		if err != nil {
			return nil, err
		}

		paths = append(paths, modulePath)
	}

	return paths, nil
}

type ModulesJson struct {
	Modules []ModuleMetadata `json:"Modules"`
}

type ModuleMetadata struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}
