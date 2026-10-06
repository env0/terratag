package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExitCodes tests the exit codes of the terratag CLI.
// It builds the binary and runs it as a subprocess to verify exit behavior.
func TestExitCodes(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(t.TempDir(), "terratag")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	tests := []struct {
		name         string
		args         []string
		env          []string
		expectedCode int
		output       []string
	}{
		{
			name:         "missing tags returns exit code 2",
			args:         []string{},
			expectedCode: 2,
		},
		{
			name:         "invalid type returns exit code 2",
			args:         []string{"-tags={\"key\":\"value\"}", "-type=invalid"},
			expectedCode: 2,
		},
		{
			name:         "version flag returns exit code 0",
			args:         []string{"-version"},
			expectedCode: 0,
		},
		{
			name:         "non-existent directory returns exit code 1",
			args:         []string{"-tags={\"key\":\"value\"}", "-dir=/nonexistent/path/that/does/not/exist"},
			expectedCode: 1,
		},
		{
			name:         "help flag returns exit code 0 and mentions .tofu files",
			args:         []string{"-h"},
			expectedCode: 0,
			output:       []string{".tf and .tofu files", "<basename>.terratag.tf (or .terratag.tofu)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Env = append(os.Environ(), tt.env...)

			out, err := cmd.CombinedOutput()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("unexpected error type: %v", err)
				}
			}

			if exitCode != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, exitCode)
			}

			for _, want := range tt.output {
				if !strings.Contains(string(out), want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, out)
				}
			}
		})
	}
}
