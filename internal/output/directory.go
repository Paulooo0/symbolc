package output

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func WriteDirectoryAtomic(root string, files map[string][]byte) error {
	parent := filepath.Dir(root)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("output: create parent directory: %w", err)
	}

	temporary, err := os.MkdirTemp(parent, ".symbolc-new-*")
	if err != nil {
		return fmt.Errorf("output: create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporary)

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		destination := filepath.Join(temporary, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("output: create artifact directory: %w", err)
		}
		if err := os.WriteFile(destination, files[path], 0o644); err != nil {
			return fmt.Errorf("output: write artifact: %w", err)
		}
	}

	backup, err := os.CreateTemp(parent, ".symbolc-old-*")
	if err != nil {
		return fmt.Errorf("output: create backup path: %w", err)
	}
	backupPath := backup.Name()
	if err := backup.Close(); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("output: close backup path: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("output: remove backup placeholder: %w", err)
	}
	defer os.RemoveAll(backupPath)

	hadOutput := false
	if _, err := os.Lstat(root); err == nil {
		if err := os.Rename(root, backupPath); err != nil {
			return fmt.Errorf("output: move previous output: %w", err)
		}
		hadOutput = true
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("output: inspect previous output: %w", err)
	}

	if err := os.Rename(temporary, root); err != nil {
		if hadOutput {
			_ = os.Rename(backupPath, root)
		}
		return fmt.Errorf("output: publish output: %w", err)
	}
	if hadOutput {
		if err := os.RemoveAll(backupPath); err != nil {
			return fmt.Errorf("output: remove previous output: %w", err)
		}
	}
	return nil
}
