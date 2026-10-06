package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/env0/terratag/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilePathsIncludesTofuFiles(t *testing.T) {
	for _, iacType := range []common.IACType{common.Terraform, common.Terragrunt, common.TerragruntRunAll} {
		t.Run(string(iacType), func(t *testing.T) {
			// t.TempDir() can sit behind a symlink (/var on macOS); terraform mode returns resolved paths.
			dir, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)

			for _, name := range []string{"main.tf", "main.tofu", "main.tofu.bak", "terragrunt.hcl"} {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
			}

			got, err := GetFilePaths(dir, string(iacType))
			require.NoError(t, err)

			assert.ElementsMatch(t, []string{filepath.Join(dir, "main.tf"), filepath.Join(dir, "main.tofu")}, got)
		})
	}
}

func TestGetFilePathsSkipsDirectoriesInTerragrunt(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "module.tofu"), 0o700))

	got, err := GetFilePaths(dir, string(common.Terragrunt))
	require.NoError(t, err)

	assert.Empty(t, got)
}
