package terraform

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	version "github.com/hashicorp/go-version"
)

var (
	terragruntVersionOnce sync.Once
	terragruntVersion     string
	terragruntParsed      *version.Version
	terragruntVersionErr  error
)

var terragruntVersionRegex = regexp.MustCompile(`\d+\.\d+\.\d+`)

const TerragruntRunMinVersion = "0.78.0"

func GetTerragruntVersion() (string, error) {
	terragruntVersionOnce.Do(func() {
		cmd := exec.Command("terragrunt", "--version")
		out, err := cmd.Output()
		if err != nil {
			terragruntVersionErr = fmt.Errorf("failed to run 'terragrunt version': %w", err)
			return
		}

		match := terragruntVersionRegex.FindStringSubmatch(string(out))
		if len(match) < 1 {
			terragruntVersionErr = fmt.Errorf("failed to parse terragrunt version from output: %s", strings.TrimSpace(string(out)))
			return
		}

		terragruntVersion = match[0]
		terragruntParsed, terragruntVersionErr = version.NewVersion(terragruntVersion)
	})

	return terragruntVersion, terragruntVersionErr
}

func IsTerragruntVersionAtLeast(minVersion string) (bool, error) {
	if _, err := GetTerragruntVersion(); err != nil {
		return false, err
	}

	minVer, err := version.NewVersion(minVersion)
	if err != nil {
		return false, fmt.Errorf("invalid min terragrunt version '%s': %w", minVersion, err)
	}

	return terragruntParsed.GreaterThanOrEqual(minVer), nil
}

func IsTerragruntRunSupported() (bool, error) {
	return IsTerragruntVersionAtLeast(TerragruntRunMinVersion)
}
