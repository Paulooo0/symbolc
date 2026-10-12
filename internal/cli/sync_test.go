package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncCommand(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "symbolc.yml")
	root, err := filepath.Abs(filepath.Join("..", "..", "internal", "analyzer", "testdata", "basic"))
	require.NoError(t, err)
	configData := []byte(fmt.Sprintf("version: 1\nsource:\n  roots:\n    - %s\n", root))
	require.NoError(t, os.WriteFile(configPath, configData, 0o600))

	var output bytes.Buffer
	previousOutput := RootCmd.OutOrStdout()
	defer RootCmd.SetOut(previousOutput)
	RootCmd.SetOut(&output)
	RootCmd.SetArgs([]string{"sync", "--config", configPath})
	defer RootCmd.SetArgs(nil)

	require.NoError(t, RootCmd.Execute())
	require.Contains(t, output.String(), "version: 1")
	require.Contains(t, output.String(), "executables:")
	require.NotContains(t, output.String(), "configuration loaded")
}

func TestSyncCommandWritesConfiguredOutput(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "symbolc.yml")
	outputPath := filepath.Join(directory, "context.yml", "index.yml")
	root, err := filepath.Abs(filepath.Join("..", "..", "internal", "analyzer", "testdata", "basic"))
	require.NoError(t, err)
	configData := []byte(fmt.Sprintf(`version: 1
source:
  roots:
    - %s
output:
  path: context.yml
`, root))
	require.NoError(t, os.WriteFile(configPath, configData, 0o600))

	var stdout bytes.Buffer
	previousOutput := RootCmd.OutOrStdout()
	defer RootCmd.SetOut(previousOutput)
	RootCmd.SetOut(&stdout)
	RootCmd.SetArgs([]string{"sync", "--config", configPath})
	defer RootCmd.SetArgs(nil)

	require.NoError(t, RootCmd.Execute())
	require.Empty(t, stdout.String())

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "version: 1")
}
