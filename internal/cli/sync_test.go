package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestSyncCommandPreservesOutputWhenAnalysisFails(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "symbolc.yml")
	outputPath := filepath.Join(directory, "context")
	root, err := filepath.Abs(filepath.Join("..", "..", "internal", "analyzer", "testdata", "basic"))
	require.NoError(t, err)

	writeConfig := func(sourceRoot string) {
		data := []byte(fmt.Sprintf(`version: 1
source:
  roots:
    - %s
output:
  path: %s
`, sourceRoot, outputPath))
		require.NoError(t, os.WriteFile(configPath, data, 0o600))
	}

	writeConfig(root)
	RootCmd.SetArgs([]string{"sync", "--config", configPath})
	require.NoError(t, RootCmd.Execute())

	indexPath := filepath.Join(outputPath, "index.yml")
	previous, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	writeConfig(filepath.Join(directory, "does-not-exist"))
	err = RootCmd.Execute()
	require.Error(t, err)

	current, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	require.Equal(t, previous, current)
}

func TestSyncCommandRemovesFilteredArtifacts(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "symbolc.yml")
	outputPath := filepath.Join(directory, "context")
	root, err := filepath.Abs(filepath.Join("..", "..", "internal", "analyzer", "testdata", "basic"))
	require.NoError(t, err)

	writeConfig := func(exclude string) {
		data := fmt.Sprintf(`version: 1
source:
  roots:
    - %s
contracts:
  exclude:
    - "%s"
output:
  path: %s
`, root, exclude, outputPath)
		require.NoError(t, os.WriteFile(configPath, []byte(data), 0o600))
	}

	writeConfig("never-match")
	RootCmd.SetArgs([]string{"sync", "--config", configPath})
	require.NoError(t, RootCmd.Execute())

	contracts := filepath.Join(outputPath, "example.com_testproject", "contracts")
	entries, err := os.ReadDir(contracts)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	writeConfig("**/*.go")
	require.NoError(t, RootCmd.Execute())

	entries, err = os.ReadDir(contracts)
	require.ErrorIs(t, err, os.ErrNotExist)
	require.Nil(t, entries)

	var stale bool
	require.NoError(t, filepath.WalkDir(outputPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.Contains(filepath.ToSlash(path), "/contracts/") {
			stale = true
		}
		return nil
	}))
	require.False(t, stale)
}
