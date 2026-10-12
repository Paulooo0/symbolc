package output

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteFileAtomic(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("output: create directory: %w", err)
	}

	temporary, err := os.CreateTemp(directory, ".symbolc-*")
	if err != nil {
		return fmt.Errorf("output: create temporary file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)

	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("output: set file permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("output: write temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("output: close temporary file: %w", err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("output: replace file: %w", err)
	}
	return nil
}
