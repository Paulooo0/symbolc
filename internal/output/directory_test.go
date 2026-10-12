package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteDirectoryAtomicReplacesManagedOutput(t *testing.T) {
	root := filepath.Join(t.TempDir(), "context")
	require.NoError(t, WriteDirectoryAtomic(root, map[string][]byte{
		"index.yml":                []byte("version: 1\n"),
		"contracts/old.yml":        []byte("old\n"),
		"flux-entrypoints/app.yml": []byte("app\n"),
	}))

	require.NoError(t, WriteDirectoryAtomic(root, map[string][]byte{
		"index.yml":         []byte("version: 2\n"),
		"contracts/new.yml": []byte("new\n"),
	}))

	data, err := os.ReadFile(filepath.Join(root, "index.yml"))
	require.NoError(t, err)
	require.Equal(t, "version: 2\n", string(data))
	_, err = os.Stat(filepath.Join(root, "contracts", "old.yml"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(root, "flux-entrypoints", "app.yml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
