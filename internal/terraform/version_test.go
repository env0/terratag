package terraform

import "testing"

func TestParseTerragruntVersion(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    string
		wantErr bool
	}{
		{
			name: "standard version output",
			out:  "terragrunt version v0.77.20",
			want: "0.77.20",
		},
		{
			name: "prerelease version",
			out:  "terragrunt version v0.88.0-alpha",
			want: "0.88.0",
		},
		{
			name:    "garbage input",
			out:     "command not found: terragrunt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTerragruntVersion(tt.out)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}
