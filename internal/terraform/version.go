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
	terragruntRunSupportedOnce sync.Once
	terragruntRunSupported     bool
	terragruntRunSupportedErr  error
)

var terragruntVersionRegex = regexp.MustCompile(`\d+\.\d+\.\d+`)

const TerragruntRunMinVersion = "0.78.0"

// parseTerragruntVersion extracts a semantic version from the output of 'terragrunt --version'.
func parseTerragruntVersion(out string) (*version.Version, error) {
	match := terragruntVersionRegex.FindString(out)
	if match == "" {
		return nil, fmt.Errorf("failed to parse terragrunt version from output: %s", strings.TrimSpace(out))
	}

	return version.NewVersion(match)
}

// IsTerragruntRunSupported reports whether the installed terragrunt version supports the 'run' command.
func IsTerragruntRunSupported() (bool, error) {
	terragruntRunSupportedOnce.Do(func() {
		cmd := exec.Command("terragrunt", "--version")

		out, err := cmd.Output()
		if err != nil {
			terragruntRunSupportedErr = fmt.Errorf("failed to run 'terragrunt --version': %w", err)
			return
		}

		parsed, err := parseTerragruntVersion(string(out))
		if err != nil {
			terragruntRunSupportedErr = err
			return
		}

		minVer, err := version.NewVersion(TerragruntRunMinVersion)
		if err != nil {
			terragruntRunSupportedErr = fmt.Errorf("invalid min terragrunt version '%s': %w", TerragruntRunMinVersion, err)
			return
		}

		terragruntRunSupported = parsed.GreaterThanOrEqual(minVer)
	})

	return terragruntRunSupported, terragruntRunSupportedErr
}
