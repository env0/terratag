package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceWithTerratagFileKeepsExtension(t *testing.T) {
	for _, ext := range []string{".tf", ".tofu"} {
		t.Run(ext, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main"+ext)
			require.NoError(t, os.WriteFile(path, []byte("original"), 0o600))

			require.NoError(t, ReplaceWithTerratagFile(path, "tagged", true))

			content, err := os.ReadFile(filepath.Join(filepath.Dir(path), "main.terratag"+ext))
			require.NoError(t, err)
			assert.Equal(t, "tagged", string(content))
		})
	}
}
