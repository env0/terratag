package main_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar"
	. "github.com/env0/terratag"
	"github.com/env0/terratag/cli"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
)

var cleanArgs = append(os.Args)
var args = append(os.Args, "-tags={\"env0_environment_id\":\"40907eff-cf7c-419a-8694-e1c6bf1d1168\",\"env0_project_id\":\"43fd4ff1-8d37-4d9d-ac97-295bd850bf94\"}")
var rootDir = "test/fixture"

type TestCase struct {
	suite    string
	suiteDir string
	entryDir string
}

func TestTerraform11(t *testing.T) {
	testTerraform(t, "11")
}

func TestTerraform12(t *testing.T) {
	testTerraform(t, "12")
}

func TestTerraform13(t *testing.T) {
	testTerraform(t, "13_and_above")
}

func TestTerraform14(t *testing.T) {
	testTerraform(t, "13_and_above")
}

func TestTerraform15(t *testing.T) {
	testTerraform(t, "13_and_above")
}

func testTerraform(t *testing.T, version string) {
	for _, tt := range getEntries(version) {
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

	g.Expect(len(actualTerratag)).To(BeEquivalentTo(len(expectedTerratag)), "it should generate the same number of terratag files as expected")
	for _, expectedTerratagFile := range expectedTerratag {
		expectedFile, _ := os.Open(expectedTerratagFile)
		expectedContent, _ := ioutil.ReadAll(expectedFile)
		actualTerratagFile := strings.ReplaceAll(expectedTerratagFile, "/expected/", "/out/")
		actualFile, _ := os.Open(actualTerratagFile)
		actualContent, _ := ioutil.ReadAll(actualFile)
		g.Expect(string(expectedContent)).To(BeEquivalentTo(string(actualContent)), actualTerratagFile+" does not match "+expectedTerratagFile)
	}
}

func itShouldRunTerraformValidate(entryDir string, g *GomegaWithT) {
	err := terraform(entryDir, "validate")
	g.Expect(err).To(BeNil(), "terraform validate failed")
}

func itShouldRunTerratag(entryDir string, g *GomegaWithT) {
	err := terratag(entryDir)
	g.Expect(err).To(BeNil(), "terratag failed")
}

func itShouldTerraformInit(entryDir string, g *GomegaWithT) {
	err := terraform(entryDir, "init")
	g.Expect(err).To(BeNil(), "terraform init failed")
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
		terraformPathSplit := strings.Split(entryFile, terraformDir)
		pathAfterTerraformDir := terraformPathSplit[1]
		suite := strings.Split(pathAfterTerraformDir, "/")[1]
		pathBeforeTerraformDir := terraformPathSplit[0]
		suiteDir := pathBeforeTerraformDir + terraformDir + "/" + suite

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
	args, isMissingArg := cli.InitArgs()
	if isMissingArg {
		return errors.New("Missing arg")
	}
	Terratag(args)
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
