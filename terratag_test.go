package main_test

import (
	"fmt"
	"github.com/bmatcuk/doublestar"
	. "github.com/env0/terratag"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const rootDir = "test/fixture"

var cleanArgs = append(os.Args)
var testEntries []table.TableEntry

//terraform12Entries := getEntries("12")

var _ = Describe("Terratag", func() {
	SynchronizedBeforeSuite(func() []byte {
		testEntries = getEntries("11")
		return nil
	}, func(_ []byte) {

	})

	describeTerraform(testEntries, "11")
})

func describeTerraform(testEntries []table.TableEntry, version string) {
	table.DescribeTable("Terraform "+version,
		func(entryDir string, suiteDir string) {
			itShouldTerraformInit(entryDir)
			itShouldRunTerratag(entryDir)
			itShouldRunTerraformValidate(entryDir)
			itShouldGenerateExpectedTerratagFiles(suiteDir)
		}, testEntries...,
	)
}

func itShouldGenerateExpectedTerratagFiles(suiteDir string) {
	expectedPattern := suiteDir + "/expected/**/*.terratag.tf"
	var expectedTerratag []string
	var actualTerratag []string
	expectedTerratag, _ = doublestar.Glob(expectedPattern)
	actualTerratag, _ = doublestar.Glob(suiteDir + "/out/**/*.terratag.tf")
	actualTerratag = filterSymlink(actualTerratag)

	Expect(len(actualTerratag)).To(BeEquivalentTo(len(expectedTerratag)))
	for _, expectedTerratagFile := range expectedTerratag {
		expectedFile, _ := os.Open(expectedTerratagFile)
		expectedContent, _ := ioutil.ReadAll(expectedFile)
		actualTerratagFile := strings.ReplaceAll(expectedTerratagFile, "/expected/", "/out/")
		actualFile, _ := os.Open(actualTerratagFile)
		actualContent, _ := ioutil.ReadAll(actualFile)
		Expect(string(expectedContent)).To(BeEquivalentTo(string(actualContent)))
	}
}

func itShouldRunTerraformValidate(entryDir string) {
	err := terraform(entryDir, "validate")
	Expect(err).To(BeNil())
}

func itShouldRunTerratag(entryDir string) {
	err := terratag(entryDir)
	Expect(err).To(BeNil())
}

func itShouldTerraformInit(entryDir string) {
	err := terraform(entryDir, "init")
	Expect(err).To(BeNil())
}

func getEntries(version string) []table.TableEntry {
	terraformDir := "/terraform_" + version
	const inputDirsMatcher = "/**/input/"
	inputDirs, _ := doublestar.Glob(rootDir + terraformDir + inputDirsMatcher)
	cloneOutput(inputDirs)

	const entryFilesMatcher = "/**/out/**/main.tf"
	entryFiles, _ := doublestar.Glob(rootDir + terraformDir + entryFilesMatcher)
	var testEntries []table.TableEntry
	for _, entryFile := range entryFiles {
		entryDir := strings.TrimSuffix(entryFile, "/main.tf")
		suite := strings.Split(strings.Split(entryFile, terraformDir)[1], "/")[1]
		suiteDir := strings.Split(entryFile, terraformDir)[0] + terraformDir + "/" + suite

		testEntries = append(testEntries, table.Entry(suite, entryDir, suiteDir))
	}
	return testEntries
}

func terraformDir(version string) string {
	return "/terraform_" + version
}

func cloneOutput(inputDirs []string) {
	for _, inputDir := range inputDirs {
		outputDir := strings.TrimSuffix(inputDir, "input") + "out"
		os.RemoveAll(outputDir)
		copy.Copy(inputDir, outputDir)
	}
}

func terratag(entryDir string) (err interface{}) {
	defer func() {
		if innerErr := recover(); innerErr != nil {
			fmt.Println(innerErr)
			err = innerErr
		}
	}()
	var args = append(cleanArgs, "-tags={\"env0_environment_id\":\"40907eff-cf7c-419a-8694-e1c6bf1d1168\",\"env0_project_id\":\"43fd4ff1-8d37-4d9d-ac97-295bd850bf94\"}", "-dir="+entryDir)
	os.Args = args
	Terratag()
	os.Args = cleanArgs

	return nil
}

func terraform(entryDir string, cmd string) error {
	println("terraform", cmd)
	command := exec.Command("terraform", cmd)
	command.Dir = entryDir
	output, err := command.Output()
	println(string(output))

	return err
}

func filterSymlink(ss []string) (ret []string) {
	for _, s := range ss {
		resolvedSymlink, _ := filepath.EvalSymlinks(s)
		if resolvedSymlink == s {
			ret = append(ret, s)
		}
	}
	return ret
}
