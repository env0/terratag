package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/env0/terratag/internal/common"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
)

func GetTerraformVersion() (*common.Version, error) {
	output, err := exec.Command("terraform", "version").Output()
	if err != nil {
		return nil, err
	}

	outputAsString := strings.TrimSpace(string(output))
	regularExpression := regexp.MustCompile(`Terraform v(\d+).(\d+)\.\d+`)
	matches := regularExpression.FindStringSubmatch(outputAsString)[1:]

	if matches == nil {
		return nil, errors.New("unable to parse 'terraform version'")
	}

	majorVersion, err := getVersionPart(matches, Major)
	if err != nil {
		return nil, err
	}
	minorVersion, err := getVersionPart(matches, Minor)
	if err != nil {
		return nil, err
	}

	if (majorVersion == 0 && minorVersion < 11 || minorVersion > 15) || (majorVersion == 1 && minorVersion > 2) {
		return nil, fmt.Errorf("terratag only supports Terraform from version 0.11.x and up to 1.2.x - your version says %s", outputAsString)
	}

	return &common.Version{Major: majorVersion, Minor: minorVersion}, nil
}

type VersionPart int

const (
	Major VersionPart = iota
	Minor
)

func (w VersionPart) EnumIndex() int {
	return int(w)
}

func getVersionPart(parts []string, versionPart VersionPart) (int, error) {
	version, err := strconv.Atoi(parts[versionPart])
	if err != nil {
		return -1, fmt.Errorf("unable to parse %s as integer", parts[versionPart])
	}

	return version, nil
}

func GetResourceType(resource hclwrite.Block) string {
	return resource.Labels()[0]
}

func getRootDir(iacType string) string {
	if iacType == string(common.Terragrunt) {
		if _, err := os.Stat("/.terragrunt-cache"); err != nil {
			if os.IsNotExist(err) {
				return "/"
			}
		}
		return "/.terragrunt-cache"
	} else {
		return "/.terraform"
	}
}

func ValidateInitRun(dir string, iacType string) error {
	path := dir + getRootDir(iacType)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s init must run before running terratag", iacType)
		}

		return fmt.Errorf("couldn't determine if %s init has run: %v", iacType, err)
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
	rootDir += getRootDir(string(common.Terragrunt))

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
	var paths []string
	var modulesJson ModulesJson

	jsonFile, err := os.Open(dir + "/.terraform/modules/modules.json")
	//lint:ignore SA5001 not required to check file close status.
	defer jsonFile.Close()

	if os.IsNotExist(err) {
		return paths, nil
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
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
