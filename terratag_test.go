package terratag

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/bmatcuk/doublestar"
	"github.com/env0/terratag/cli"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var cleanArgs = append([]string(nil), os.Args...)
var programName = os.Args[0]
var args = []string{
	programName,
	"-tags={\"env0_environment_id\":\"40907eff-cf7c-419a-8694-e1c6bf1d1168\",\"env0_project_id\":\"43fd4ff1-8d37-4d9d-ac97-295bd850bf94\"}",
}
var testsDir = "test/tests"
var fixtureDir = "test/fixture"
var osArgsLock sync.Mutex

type TestCase struct {
	suite    string
	suiteDir string
	entryDir string
}

type TestCaseConfig struct {
	Suites []string
}

func TestTerraform11(t *testing.T) {
	testTerraform(t, "11")
}

func TestTerraform12(t *testing.T) {
	testTerraform(t, "12")
}

func TestTerraform13(t *testing.T) {
	testTerraform(t, "13_14")
}

func TestTerraform14(t *testing.T) {
	testTerraform(t, "13_14")
}

func TestTerraform15(t *testing.T) {
	testTerraform(t, "15_1.0")
}

func TestTerraform1o0(t *testing.T) {
	testTerraform(t, "15_1.0")
}

func TestTerraform1o0WithFilter(t *testing.T) {
	testTerraformWithFilter(t, "15_1.0_filter", false, "azurerm_resource_group|aws_s3_bucket")
}

func TestTerraform1o1(t *testing.T) {
	testTerraform(t, "15_1.1")
}

func TestTerraform1o1WithFilter(t *testing.T) {
	testTerraformWithFilter(t, "15_1.1_filter", false, "azurerm_resource_group|aws_s3_bucket")
}

func TestTerraform1o2(t *testing.T) {
	testTerraform(t, "15_1.2")
}

func TestTerraform1o2WithFilter(t *testing.T) {
	testTerraformWithFilter(t, "15_1.2_filter", false, "azurerm_resource_group|aws_s3_bucket")
}

func TestTerraform1o2WithInvertedFilter(t *testing.T) {
	testTerraformWithFilter(t, "15_1.2_inverted_filter", true, "azurerm_resource_group")
}

func TestTerragruntWithCache(t *testing.T) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	g := NewWithT(t)

	entryDir := "./test/tests/terragrunt_with_cache"

	in := entryDir + "/in"
	out := entryDir + "/out"

	if err := os.RemoveAll(out); err != nil {
		t.Fatalf("failed to remove out directory: %s %s", out, err.Error())
	}

	if err := copy.Copy(in, out); err != nil {
		t.Fatalf("failed to in directory to out directory: %s", err.Error())
	}

	itShouldRunTerragruntInit(out, g)
	itShouldRunTerratagTerragruntMode(out, g)
	itShouldRunTerragruntValidate(out, g)
	itShouldGenerateExpectedTerragruntTerratagFiles(entryDir, g)
}

func testTerraform(t *testing.T, version string) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	entries := getEntries(t, version)
	if len(entries) == 0 {
		t.Fatalf("test entries not found for version %s", version)
	}

	for _, tt := range entries {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.suite, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			g := NewWithT(t)
			itShouldTerraformInit(tt.entryDir, g)
			itShouldRunTerratag(tt.entryDir, false, "", g)
			itShouldRunTerraformValidate(tt.entryDir, g)
			itShouldGenerateExpectedTerratagFiles(tt.suiteDir, g)
		})
	}
}

func testTerraformWithFilter(t *testing.T, version string, invertFilter bool, filter string) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	for _, tt := range getEntries(t, version) {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.suite, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			g := NewWithT(t)
			itShouldTerraformInit(tt.entryDir, g)
			itShouldRunTerratag(tt.entryDir, invertFilter, filter, g)
			itShouldRunTerraformValidate(tt.entryDir, g)
			itShouldGenerateExpectedTerratagFiles(tt.suiteDir, g)
		})
	}
}

func itShouldGenerateExpectedTerratagFiles(entryDir string, g *GomegaWithT) {
	expectedPattern := strings.Split(entryDir, "/out/")[0] + "/expected/*.tf"
	var expectedTerratag []string
	var actualTerratag []string
	expectedTerratag, _ = doublestar.Glob(expectedPattern)
	if len(expectedTerratag) == 0 {
		expectedPattern = strings.Split(entryDir, "/out/")[0] + "/expected/**/*.tf"
		expectedTerratag, _ = doublestar.Glob(expectedPattern)
	}
	actualTerratag, _ = doublestar.Glob(entryDir + "/*.tf")
	if len(actualTerratag) == 0 {
		actualTerratag, _ = doublestar.Glob(entryDir + "/**/*.tf")
	}
	actualTerratag = filterSymlink(actualTerratag)

	g.Expect(len(actualTerratag)).Should(BeNumerically(">", 0))
	g.Expect(len(expectedTerratag)).Should(BeNumerically(">", 0))
	g.Expect(len(actualTerratag)).To(BeEquivalentTo(len(expectedTerratag)), "it should generate the same number of terratag files as expected")
	for i, expectedTerratagFile := range expectedTerratag {
		expectedFile, _ := os.Open(expectedTerratagFile)
		expectedContent, _ := ioutil.ReadAll(expectedFile)
		actualTerratagFile := actualTerratag[i]
		actualFile, _ := os.Open(actualTerratagFile)
		actualContent, _ := ioutil.ReadAll(actualFile)
		g.Expect(string(expectedContent)).To(BeEquivalentTo(string(actualContent)), actualTerratagFile+" does not match "+expectedTerratagFile)
	}
}

func getFileSha256(filename string, g *GomegaWithT) string {
	f, err := os.Open(filename)
	g.Expect(err).To(BeNil())
	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	g.Expect(err).To(BeNil())

	return string(h.Sum(nil))
}

func itShouldGenerateExpectedTerragruntTerratagFiles(entryDir string, g *GomegaWithT) {
	expectedPattern := entryDir + "/expected/**/*.tf"
	expectedTerratag, _ := doublestar.Glob(expectedPattern)

	actualTerratag, _ := doublestar.Glob(entryDir + "/out/**/.terragrunt-cache/**/*.tf")
	actualTerratag = filterSymlink(actualTerratag)

	hashmap := make(map[string]string)

	for _, acctualTerratagFile := range actualTerratag {
		hashmap[getFileSha256(acctualTerratagFile, g)] = acctualTerratagFile
	}

	for _, expectedTerratagFile := range expectedTerratag {
		hash := getFileSha256(expectedTerratagFile, g)
		_, ok := hashmap[hash]
		g.Expect(ok).To(BeTrue())
	}
}

func itShouldRunTerraformValidate(entryDir string, g *GomegaWithT) {
	err := run_terraform(entryDir, "validate")
	g.Expect(err).To(BeNil(), "terraform validate failed")
}

func itShouldRunTerratag(entryDir string, invertFilter bool, filter string, g *GomegaWithT) {
	err := run_terratag(entryDir, invertFilter, filter, false)
	g.Expect(err).To(BeNil(), "terratag failed")
}

func itShouldRunTerratagTerragruntMode(entryDir string, g *GomegaWithT) {
	err := run_terratag(entryDir, false, "", true)
	g.Expect(err).To(BeNil(), "terratag terragrunt mode failed")
}

func itShouldTerraformInit(entryDir string, g *GomegaWithT) {
	err := run_terraform(entryDir, "init")
	g.Expect(err).To(BeNil(), "terraform init failed")
}

func itShouldRunTerragruntValidate(entryDir string, g *GomegaWithT) {
	err := run_terragrunt(entryDir, "validate")
	g.Expect(err).To(BeNil(), "terragrunt validate failed")
}

func itShouldRunTerragruntInit(entryDir string, g *GomegaWithT) {
	err := run_terragrunt(entryDir, "init")
	g.Expect(err).To(BeNil(), "terragrunt init failed")
}

func getConfig(terraformDir string) (*TestCaseConfig, error) {
	var config TestCaseConfig

	viper.SetConfigType("yaml")
	viper.SetConfigFile(fixtureDir + terraformDir + "/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func getEntries(t *testing.T, version string) []TestCase {
	terraformDir := "/terraform_" + version

	config, err := getConfig(terraformDir)
	if err != nil {
		t.Fatalf("failed to load test case config for version %s: %v", version, err)
	}
	suitesMap := make(map[string]interface{})
	for _, suite := range config.Suites {
		suitesMap[suite] = nil
	}

	const inputDirsMatcher = "/**/input/"
	inputDirs, _ := doublestar.Glob(testsDir + inputDirsMatcher)
	cloneOutput(inputDirs, terraformDir)

	entryFilesMatcher := "/**/out" + terraformDir + "/**/main.tf"
	entryFiles, _ := doublestar.Glob(testsDir + entryFilesMatcher)

	var testEntries []TestCase
	for _, entryFile := range entryFiles {
		// convert windows paths to use forward slashes
		slashed := filepath.ToSlash(entryFile)
		entryDir := strings.TrimSuffix(slashed, "/main.tf")
		terraformPathSplit := strings.Split(slashed, terraformDir)
		pathBeforeTerraformDir := terraformPathSplit[0]
		suiteDir := pathBeforeTerraformDir + terraformDir
		suite := strings.Split(pathBeforeTerraformDir, "/")[2]

		if _, ok := suitesMap[suite]; !ok {
			// Not in configuration file. Skip test.
			continue
		}

		testEntries = append(testEntries, TestCase{
			suite:    suite,
			suiteDir: suiteDir,
			entryDir: entryDir,
		})
	}

	return testEntries
}

func cloneOutput(inputDirs []string, terraformDir string) {
	for _, inputDir := range inputDirs {
		outputDir := strings.TrimSuffix(inputDir, "input") + "out" + terraformDir
		os.RemoveAll(outputDir)
		copy.Copy(inputDir, outputDir)
	}
}

func run_terratag(entryDir string, invertFilter bool, filter string, terragrunt bool) (err interface{}) {
	defer func() {
		if innerErr := recover(); innerErr != nil {
			fmt.Println(innerErr)
			err = innerErr
		}
	}()
	osArgsLock.Lock()

	os.Args = append(args, "-dir="+entryDir)

	if invertFilter {
		os.Args = append(os.Args, "--invertFilter="+strconv.FormatBool(invertFilter))
	}

	if filter != "" {
		os.Args = append(os.Args, "-filter="+filter)
	}

	if terragrunt {
		os.Args = append(os.Args, "-type=terragrunt", "-rename=false")
	}

	args, err := cli.InitArgs()
	os.Args = cleanArgs
	osArgsLock.Unlock()
	if err != nil {
		return err
	}
	Terratag(args)

	return nil
}

func run(prog string, entryDir string, cmd string) error {
	println(prog, cmd)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := exec.Command(prog, cmd)
	command.Dir = entryDir
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		log.Println(stderr.String())
		return err
	}

	println(stdout.String())

	return nil
}

func run_terraform(entryDir string, cmd string) error {
	return run("terraform", entryDir, cmd)
}

func run_terragrunt(entryDir string, cmd string) error {
	return run("terragrunt", entryDir, cmd)
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

func TestToHclMap(t *testing.T) {
	validCases := map[string]string{
		`{"a":"b","c":"d"}`: `{"a"="b","c"="d"}`,
		`a=b,c=d`:           `{"a"="b","c"="d"}`,
		`a-key=b-value`:     `{"a-key"="b-value"}`,
		"{}":                "{}",
	}

	for input, output := range validCases {
		input, expectedOutput := input, output
		t.Run("valid input "+input, func(t *testing.T) {
			output, err := toHclMap(input)
			assert.NoError(t, err)
			assert.Equal(t, expectedOutput, output)
		})
	}

	invalidCases := []string{
		"a$#$=b",
		`{"a": {"b": "c"}}`,
		"_a=b",
		"5a=b",
		"a=b!",
	}

	for i := range invalidCases {
		input := invalidCases[i]
		t.Run("invalid input "+input, func(t *testing.T) {
			_, err := toHclMap(input)
			assert.Error(t, err)
		})
	}
}
