package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Paulooo0/symbolc/pkg/ptr"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		contents    string
		wantVersion *int8
		wantRoot    string
		wantErr     string
	}{
		{
			name: "loads valid configuration",
			contents: `version: 1
source:
  roots:
    - ./internal
contracts:
  include:
    - internal/**
  exclude:
    - "**/*_test.go"
`,
			wantVersion: ptr.ToPtr(int8(1)),
			wantRoot:    "./internal",
		},
		{
			name: "fallbacks empty version",
			contents: `
source:
  roots:
    - ./internal
contracts:
  include:
    - internal/**
  exclude:
    - "**/*_test.go"
`,
			wantVersion: ptr.ToPtr(int8(1)),
			wantRoot:    "./internal",
		},
		{
			name:     "rejects malformed yaml",
			contents: "version: [",
			wantErr:  "config: parse yaml data:",
		},
		{
			name: "rejects unsupported version",
			contents: `version: 2
source:
  roots:
    - .
`,
			wantErr: "config: invalid config: config: unsupported config version: 2",
		},
		{
			name:     "rejects missing roots",
			contents: "version: 1\n",
			wantErr:  "config: invalid config: config: source.roots must contain at least one path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "symbolc.yml")
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o600))

			got, err := Load(path)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tt.wantErr), err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.wantVersion, got.Version)
			require.Equal(t, []string{tt.wantRoot}, got.Source.Roots)
			require.Equal(t, []string{"internal/**"}, got.Contracts.Include)
			require.Equal(t, []string{"**/*_test.go"}, got.Contracts.Exclude)
		})
	}
}
