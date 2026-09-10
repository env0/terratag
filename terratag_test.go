package terratag

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/bmatcuk/doublestar"
	"github.com/env0/terratag/cli"
	"github.com/env0/terratag/internal/common"
	"github.com/env0/terratag/internal/terraform"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	testTerraform(t, "15")
}

func TestTerraformlatestWithFilter(t *testing.T) {
	testTerraformWithFilter(t, "latest_filter", "azurerm_resource_group|aws_s3_bucket", "")
}

func TestTerraformlatestWithSkip(t *testing.T) {
	testTerraformWithFilter(t, "latest_skip", ".*", "azurerm_resource_group")
}

func TestTerraformlatest(t *testing.T) {
	testTerraform(t, "latest")
}

func TestOpenTofu(t *testing.T) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	version := "latest"

	entries := getEntries(t, version)
	if len(entries) == 0 {
		t.Fatalf("test entries not found for version %s", version)
	}

	for _, tt := range entries {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.suite, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			g := NewWithT(t)
			itShouldOpenTofuInit(tt.entryDir, g)
			itShouldRunTerratag(tt.entryDir, "", "", g)
			itShouldRunOpenTofuValidate(tt.entryDir, g)
			itShouldGenerateExpectedTerratagFiles(tt.suiteDir, g)
		})
	}
}

func TestTerragruntWithCache(t *testing.T) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	g := NewWithT(t)

	entryDir := "./test/tests/terragrunt_unit"

	in := entryDir + "/in"
	out := entryDir + "/out"
	unit := out + "/unit"

	if err := os.RemoveAll(out); err != nil {
		t.Fatalf("failed to remove out directory: %s %s", out, err.Error())
	}

	if err := copy.Copy(in, out); err != nil {
		t.Fatalf("failed to in directory to out directory: %s", err.Error())
	}

	itShouldRunTerragruntInit(unit, g)
	itShouldRunTerratagTerragruntMode(unit, g)
	itShouldRunTerragruntValidate(unit, g)
	itShouldGenerateExpectedTerragruntTerratagFiles(entryDir, g)
}

func TestTerragruntRunAll(t *testing.T) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	g := NewWithT(t)

	entryDir := "./test/tests/terragrunt_implicit_stack"

	in := entryDir + "/in"
	out := entryDir + "/out"
	stack := out + "/stack"

	if err := os.RemoveAll(out); err != nil {
		t.Fatalf("failed to remove out directory: %s %s", out, err.Error())
	}

	if err := copy.Copy(in, out); err != nil {
		t.Fatalf("failed to copy in directory to out directory: %s", err.Error())
	}

	itShouldRunTerragruntRunAllInit(stack, g)
	itShouldRunTerratagTerragruntRunAllMode(stack, g)
	itShouldRunTerragruntRunAllValidate(stack, g)
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
			itShouldRunTerratag(tt.entryDir, "", "", g)
			itShouldRunTerraformValidate(tt.entryDir, g)
			itShouldGenerateExpectedTerratagFiles(tt.suiteDir, g)
		})
	}
}

func testTerraformWithFilter(t *testing.T, version string, filter string, skip string) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
	}

	for _, tt := range getEntries(t, version) {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.suite, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			g := NewWithT(t)
			itShouldTerraformInit(tt.entryDir, g)
			itShouldRunTerratag(tt.entryDir, filter, skip, g)
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
		expectedContent, _ := io.ReadAll(expectedFile)
		actualTerratagFile := actualTerratag[i]
		actualFile, _ := os.Open(actualTerratagFile)
		actualContent, _ := io.ReadAll(actualFile)
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

	cachePattern := entryDir + "/out/**/unit*/.terragrunt-cache"
	cacheDirs, _ := doublestar.Glob(cachePattern)

	g.Expect(len(cacheDirs)).To(BeNumerically(">", 0))

	for _, cacheDir := range cacheDirs {
		hashmap := make(map[string]string)

		actualTerratag, _ := doublestar.Glob(cacheDir + "/**/*.tf")
		actualTerratag = filterSymlink(actualTerratag)

		for _, actualFile := range actualTerratag {
			hash := getFileSha256(actualFile, g)
			hashmap[hash] = actualFile
		}

		for _, expectedTerratagFile := range expectedTerratag {
			hash := getFileSha256(expectedTerratagFile, g)
			_, ok := hashmap[hash]
			g.Expect(ok).To(BeTrue(), "expected file "+expectedTerratagFile+" not found in cache directory "+cacheDir)
		}
	}
}

func itShouldRunOpenTofuValidate(entryDir string, g *GomegaWithT) {
	err := run_opentofu(entryDir, "validate")
	g.Expect(err).To(BeNil(), "opentofu validate failed")
}

func itShouldRunTerraformValidate(entryDir string, g *GomegaWithT) {
	err := run_terraform(entryDir, "validate")
	g.Expect(err).To(BeNil(), "terraform validate failed")
}

func itShouldRunTerratag(entryDir string, filter string, skip string, g *GomegaWithT) {
	err := run_terratag(entryDir, filter, skip, common.Terraform)
	g.Expect(err).To(BeNil(), "terratag failed")
}

func itShouldRunTerratagTerragruntMode(entryDir string, g *GomegaWithT) {
	err := run_terratag(entryDir, "", "", common.Terragrunt)
	g.Expect(err).To(BeNil(), "terratag terragrunt mode failed")
}

func itShouldRunTerratagTerragruntRunAllMode(entryDir string, g *GomegaWithT) {
	err := run_terratag(entryDir, "", "", common.TerragruntRunAll)
	g.Expect(err).To(BeNil(), "terratag terragrunt run-all mode failed")
}

func itShouldOpenTofuInit(entryDir string, g *GomegaWithT) {
	err := run_opentofu(entryDir, "init")
	g.Expect(err).To(BeNil(), "opentofu init failed")
}

func itShouldTerraformInit(entryDir string, g *GomegaWithT) {
	err := run_terraform(entryDir, "init")
	g.Expect(err).To(BeNil(), "terraform init failed")
}

func itShouldRunTerragruntValidate(entryDir string, g *GomegaWithT) {
	err := run_terragrunt(entryDir, "validate", false)
	g.Expect(err).To(BeNil(), "terragrunt validate failed")
}

func itShouldRunTerragruntInit(entryDir string, g *GomegaWithT) {
	err := run_terragrunt(entryDir, "init", false)
	g.Expect(err).To(BeNil(), "terragrunt init failed")
}

func itShouldRunTerragruntRunAllValidate(entryDir string, g *GomegaWithT) {
	err := run_terragrunt(entryDir, "validate", true)
	g.Expect(err).To(BeNil(), "terragrunt run-all validate failed")
}

func itShouldRunTerragruntRunAllInit(entryDir string, g *GomegaWithT) {
	err := run_terragrunt(entryDir, "init", true)
	g.Expect(err).To(BeNil(), "terragrunt run-all init failed")
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

	testEntries := []TestCase{}

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

func run_terratag(entryDir string, filter string, skip string, iacType common.IACType) (err any) {
	defer func() {
		if innerErr := recover(); innerErr != nil {
			fmt.Println(innerErr)
			err = innerErr
		}
	}()
	osArgsLock.Lock()

	os.Args = append(args, "-dir="+entryDir)

	if filter != "" {
		os.Args = append(os.Args, "-filter="+filter)
	}

	if skip != "" {
		os.Args = append(os.Args, "-skip="+skip)
	}

	if iacType == common.Terragrunt {
		os.Args = append(os.Args, "-type=terragrunt", "-rename=false")
	} else if iacType == common.TerragruntRunAll {
		os.Args = append(os.Args, "-type=terragrunt-run-all", "-rename=false")
	}

	args, err := cli.InitArgs()
	os.Args = cleanArgs
	osArgsLock.Unlock()

	if err != nil {
		return err
	}

	return Terratag(args)
}

func run(prog string, entryDir string, args ...string) error {
	println(prog, strings.Join(args, " "))

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command := exec.Command(prog, args...)
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

func run_opentofu(entryDir string, cmd string) error {
	return run("tofu", entryDir, cmd)
}

func run_terragrunt(entryDir string, cmd string, runAll bool) error {
	supportsRun, err := terraform.IsTerragruntRunSupported()
	if err != nil {
		return err
	}

	args := []string{}
	if supportsRun {
		args = append(args, "run")
		if runAll {
			args = append(args, "--all")
		}
		args = append(args, cmd)
	} else {
		if runAll {
			args = append(args, "run-all", cmd)
		} else {
			args = append(args, cmd)
		}
	}

	return run("terragrunt", entryDir, args...)
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
			require.NoError(t, err)
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

func TestEnvVariables(t *testing.T) {
	os.Setenv("TERRATAG_TAGS", `{"a":"b"}`)
	os.Setenv("TERRATAG_DIR", "./dir")
	os.Setenv("TERRATAG_SKIPTERRATAGFILES", "true")
	os.Setenv("TERRATAG_FILTER", "filter")
	os.Setenv("TERRATAG_SKIP", "skip")
	os.Setenv("TERRATAG_VERBOSE", "true")
	os.Setenv("TERRATAG_RENAME", "false")
	os.Setenv("TERRATAG_TYPE", string(common.Terragrunt))

	defer func() {
		os.Unsetenv("TERRATAG_TAGS")
		os.Unsetenv("TERRATAG_DIR")
		os.Unsetenv("TERRATAG_SKIPTERRATAGFILES")
		os.Unsetenv("TERRATAG_FILTER")
		os.Unsetenv("TERRATAG_SKIP")
		os.Unsetenv("TERRATAG_VERBOSE")
		os.Unsetenv("TERRATAG_RENAME")
		os.Unsetenv("TERRATAG_TYPE")
	}()

	osArgsLock.Lock()
	defer osArgsLock.Unlock()

	os.Args = []string{programName}
	args, err := cli.InitArgs()
	os.Args = cleanArgs

	require.NoError(t, err)

	assert.Equal(t, `{"a":"b"}`, args.Tags)
	assert.Equal(t, "./dir", args.Dir)
	assert.True(t, args.IsSkipTerratagFiles)
	assert.Equal(t, "filter", args.Filter)
	assert.True(t, args.Verbose)
	assert.False(t, args.Rename)
	assert.Equal(t, string(common.Terragrunt), args.Type)

	// The command line flags have precedence over environment variables.
	os.Args = []string{programName, `-tags={"c":"d"}`}
	args, err = cli.InitArgs()
	os.Args = cleanArgs

	require.NoError(t, err)

	assert.Equal(t, `{"c":"d"}`, args.Tags)
}
