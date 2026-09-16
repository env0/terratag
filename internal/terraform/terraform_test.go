package terraform

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/env0/terratag/internal/common"
)

func TestGetFilePathsIncludesTofuFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"main.tf", "main.tofu", "ignored.hcl"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetFilePaths(dir, string(common.Terraform))
	if err != nil {
		t.Fatal(err)
	}

	for i := range got {
		got[i], err = filepath.EvalSymlinks(got[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(got)
	want := []string{filepath.Join(dir, "main.tf"), filepath.Join(dir, "main.tofu")}
	for i := range want {
		want[i], err = filepath.EvalSymlinks(want[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %q, want %q", got[i], want[i])
		}
	}
}
