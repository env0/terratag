package main_test

import (
	"fmt"
	"github.com/bmatcuk/doublestar"
	. "github.com/env0/terratag"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var cleanArgs = append(os.Args)
var args = append(os.Args, "-tags={\"env0_environment_id\":\"40907eff-cf7c-419a-8694-e1c6bf1d1168\",\"env0_project_id\":\"43fd4ff1-8d37-4d9d-ac97-295bd850bf94\"}")
var rootDir = "test/fixture"
var terraform11Entries = getEntries("11")

//terraform12Entries := getEntries("12")
type TestCase struct {
	suite    string
	suiteDir string
	entryDir string
}

func TestTerratag(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	println("installing terraform")

	command := exec.Command(".tfenv/bin/tfenv install latest:^0.11 ")
	output, err := command.Output()
	println(string(output))
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, tt := range terraform11Entries {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.suite, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			g := NewGomegaWithT(t)
			itShouldTerraformInit(tt.entryDir, g)
			itShouldRunTerratag(tt.entryDir, g)
			itShouldRunTerraformValidate(tt.entryDir, g)
			itShouldGenerateExpectedTerratagFiles(tt.suiteDir, g)
		})
	}
}

func itShouldGenerateExpectedTerratagFiles(suiteDir string, g *GomegaWithT) {
	expectedPattern := suiteDir + "/expected/**/*.terratag.tf"
	var expectedTerratag []string
	var actualTerratag []string
	expectedTerratag, _ = doublestar.Glob(expectedPattern)
	actualTerratag, _ = doublestar.Glob(suiteDir + "/out/**/*.terratag.tf")
	actualTerratag = filterSymlink(actualTerratag)

	g.Expect(len(actualTerratag)).To(BeEquivalentTo(len(expectedTerratag)))
	for _, expectedTerratagFile := range expectedTerratag {
		expectedFile, _ := os.Open(expectedTerratagFile)
		expectedContent, _ := ioutil.ReadAll(expectedFile)
		actualTerratagFile := strings.ReplaceAll(expectedTerratagFile, "/expected/", "/out/")
		actualFile, _ := os.Open(actualTerratagFile)
		actualContent, _ := ioutil.ReadAll(actualFile)
		g.Expect(string(expectedContent)).To(BeEquivalentTo(string(actualContent)))
	}
}

func itShouldRunTerraformValidate(entryDir string, g *GomegaWithT) {
	err := terraform(entryDir, "validate")
	g.Expect(err).To(BeNil())
}

func itShouldRunTerratag(entryDir string, g *GomegaWithT) {
	err := terratag(entryDir)
	g.Expect(err).To(BeNil())
}

func itShouldTerraformInit(entryDir string, g *GomegaWithT) {
	err := terraform(entryDir, "init")
	g.Expect(err).To(BeNil())
}

func getEntries(version string) []TestCase {
	terraformDir := "/terraform_" + version
	const inputDirsMatcher = "/**/input/"
	inputDirs, _ := doublestar.Glob(rootDir + terraformDir + inputDirsMatcher)
	cloneOutput(inputDirs)

	const entryFilesMatcher = "/**/out/**/main.tf"
	entryFiles, _ := doublestar.Glob(rootDir + terraformDir + entryFilesMatcher)
	var testEntries []TestCase
	for _, entryFile := range entryFiles {
		entryDir := strings.TrimSuffix(entryFile, "/main.tf")
		suite := strings.Split(strings.Split(entryFile, terraformDir)[1], "/")[1]
		suiteDir := strings.Split(entryFile, terraformDir)[0] + terraformDir + "/" + suite

		testEntries = append(testEntries, TestCase{
			suite:    suite,
			suiteDir: suiteDir,
			entryDir: entryDir,
		})
	}

	return testEntries
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
	os.Args = append(args, "-dir="+entryDir)
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
