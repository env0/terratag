package terraform

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/env0/terratag/errors"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
)

func GetTerraformVersion() int {
	output, err := exec.Command("terraform", "version").Output()
	outputAsString := strings.TrimSpace(string(output))
	errors.PanicOnError(err, &outputAsString)

	regexp.MustCompile(`Terraform v(\d+)\.(\d+)\.\d+`)
	matches := strings.Split(strings.TrimPrefix("Terraform v", strings.TrimSpace(outputAsString)), ".")
	if matches == nil {
		log.Fatalln("Unable to parse 'terraform version'")
		return -1
	}
	majorVersion := getVersionPart(matches, Major)
	minorVersion := getVersionPart(matches, Minor)

	if (majorVersion == 0 && minorVersion < 11 || minorVersion > 15) || (majorVersion == 1 && minorVersion > 0) {
		log.Fatalln("Terratag only supports Terraform from version 0.11.x and up to 1.0.x - your version says ", outputAsString)
		return -1
	}

	return minorVersion
}

type VersionPart int

const (
	Major VersionPart = iota
	Minor
)

func (w VersionPart) EnumIndex() int {
	return int(w)
}

func getVersionPart(parts []string, versionPart VersionPart) int {
	version, err := strconv.Atoi(parts[versionPart])
	if err != nil {
		log.Fatalln("Unable to parse ", parts[versionPart], "as integer")
		return -1
	}

	return version
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
	const tfFileMatcher = "/*.tf"

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

type ModulesJson struct {
	Modules []ModuleMetadata `json:"Modules"`
}

type ModuleMetadata struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}
