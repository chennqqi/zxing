package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// clean removes build artifacts (build/ and build-wasm/).
// It does not remove lib/ or wasm/ (precompiled artifacts).
func clean(args []string) error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(root, "build"),
		filepath.Join(root, "build-wasm"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("Removing %s...\n", dir)
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("failed to remove %s: %w", dir, err)
			}
		} else {
			fmt.Printf("Skipping %s (not found)\n", dir)
		}
	}

	fmt.Println("Clean complete.")
	return nil
}
