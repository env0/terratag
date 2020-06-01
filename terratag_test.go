package main_test

import (
	"fmt"
	"github.com/bmatcuk/doublestar"
	. "github.com/env0/terratag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var cleanArgs = append(os.Args)
var args = append(os.Args, "-tags={\"env0_environment_id\":\"40907eff-cf7c-419a-8694-e1c6bf1d1168\",\"env0_project_id\":\"43fd4ff1-8d37-4d9d-ac97-295bd850bf94\"}")
var rootDir = "test/fixture"

var _ = Describe("Terratag", func() {
	Describe("Terraform 11", func() {
		describeTerraform("11")
	})

	Describe("Terraform 11", func() {
		describeTerraform("12")
	})
})

func describeTerraform(version string) {
	terraformDir := "/terraform_" + version
	const inputDirsMatcher = "/**/input/"
	inputDirs, _ := doublestar.Glob(rootDir + terraformDir + inputDirsMatcher)
	cloneOutput(inputDirs)

	const entryFilesMatcher = "/**/out/**/main.tf"
	entryFiles, _ := doublestar.Glob(rootDir + terraformDir + entryFilesMatcher)
	for _, entryFile := range entryFiles {
		entryDir := strings.TrimSuffix(entryFile, "/main.tf")
		suite := strings.Split(strings.Split(entryFile, terraformDir)[1], "/")[1]
		suiteDir := strings.Split(entryFile, terraformDir)[0] + terraformDir + "/" + suite

		describeSuite(entryDir, suiteDir)
	}
}

func cloneOutput(inputDirs []string) {
	for _, inputDir := range inputDirs {
		outputDir := strings.TrimSuffix(inputDir, "input") + "out"
		os.RemoveAll(outputDir)
		copy.Copy(inputDir, outputDir)
	}
}

func describeSuite(entryDir string, suiteDir string) {
	Describe("prerequesites", func() {
		It("should terraform init", func() {
			err := terraform(entryDir, "init")
			Expect(err).To(BeNil())
		})

		It("should run terratag", func() {
			err := terratag(entryDir)
			Expect(err).To(BeNil())
		})
	})

	It("should pass terraform validation", func() {
		err := terraform(entryDir, "validate")
		Expect(err).To(BeNil())
	})

	Describe("the terratag output", func() {
		expectedPattern := suiteDir + "/expected/**/*.terratag.tf"
		var expectedTerratag []string
		var actualTerratag []string

		BeforeEach(func() {
			expectedTerratag, _ = doublestar.Glob(expectedPattern)
			actualTerratag, _ = doublestar.Glob(suiteDir + "/out/**/*.terratag.tf")
			actualTerratag = filterSymlink(actualTerratag)
		})

		It("should have exactly "+string(len(expectedTerratag))+" files tagged", func() {
			Expect(len(actualTerratag)).To(BeEquivalentTo(len(expectedTerratag)))
		})

		It("should have terratag files match expected", func() {
			for _, expectedTerratagFile := range expectedTerratag {
				expectedFile, _ := os.Open(expectedTerratagFile)
				expectedContent, _ := ioutil.ReadAll(expectedFile)
				actualTerratagFile := strings.ReplaceAll(expectedTerratagFile, "/expected/", "/out/")
				actualFile, _ := os.Open(actualTerratagFile)
				actualContent, _ := ioutil.ReadAll(actualFile)
				Expect(string(expectedContent)).To(BeEquivalentTo(string(actualContent)))
			}
		})
	})
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
