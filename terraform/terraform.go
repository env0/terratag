package terraform

import (
	"encoding/json"
	"github.com/bmatcuk/doublestar"
	"github.com/env0/terratag/errors"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetTerraformVersion() int {
	output, err := exec.Command("terraform", "version").Output()
	outputAsString := strings.TrimSpace(string(output))
	errors.PanicOnError(err, &outputAsString)

	if strings.HasPrefix(outputAsString, "Terraform v0.11") {
		return 11
	} else if strings.HasPrefix(outputAsString, "Terraform v0.12") {
		return 12
	}

	log.Fatalln("Terratag only supports Terraform 0.11.x and 0.12.x - your version says ", outputAsString)
	return -1
}

func GetResourceType(resource hclwrite.Block) string {
	return resource.Labels()[0]
}

func IsTerraformInitRun(dir string) bool {
	_, err := os.Stat(dir + "/.terraform")

	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalln("terraform init must run before running terratag")
			return false
		}

		message := "couldn't determine if terraform init has run"
		errors.PanicOnError(err, &message)
	}

	return true
}

func GetTerraformFilePaths(rootDir string) []string {
	const tfFileMatcher = "/**/*.tf"

	tfFiles, err := doublestar.Glob(rootDir + tfFileMatcher)
	errors.PanicOnError(err, nil)

	modulesDirs := getTerraformModulesDirPaths(rootDir)

	for _, moduleDir := range modulesDirs {
		matches, err := doublestar.Glob(moduleDir + tfFileMatcher)
		errors.PanicOnError(err, nil)

		tfFiles = append(tfFiles, matches...)
	}

	for i, tfFile := range tfFiles {
		resolvedTfFile, err := filepath.EvalSymlinks(tfFile)
		errors.PanicOnError(err, nil)

		tfFiles[i] = resolvedTfFile
	}

	return funk.UniqString(tfFiles)
}

func getTerraformModulesDirPaths(dir string) []string {
	var paths []string
	var modulesJson ModulesJson

	jsonFile, err := os.Open(dir + "/.terraform/modules/modules.json")

	if os.IsNotExist(err) {
		return paths
	}
	errors.PanicOnError(err, nil)

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &modulesJson)
	errors.PanicOnError(err, nil)

	for _, module := range modulesJson.Modules {
		modulePath, err := filepath.EvalSymlinks(dir + "/" + module.Dir)
		errors.PanicOnError(err, nil)

		paths = append(paths, modulePath)
	}

	err = jsonFile.Close()
	errors.PanicOnError(err, nil)

	return paths
}

func GetSingleNestedBlock(parent *hclwrite.Block, typeName string) *hclwrite.Block {
	return parent.Body().FirstMatchingBlock(typeName, nil)
}

func GetNestedBlock(parent *hclwrite.Block, typeNames []string) *hclwrite.Block {
	block := GetSingleNestedBlock(parent, typeNames[0])

	if len(typeNames) == 1 {
		return block
	} else {
		return GetNestedBlock(block, typeNames[1:])
	}
}

type ModulesJson struct {
	Modules []ModuleMetadata `json:"Modules"`
}

type ModuleMetadata struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}
