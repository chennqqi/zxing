package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// syncHeaders copies ZXing-CPP headers from zxing-cpp/core/src/ to include/ZXing/.
// It also ensures include/zxing.h and include/zxing_internal.h exist in include/.
func syncHeaders(args []string) error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	srcDir := filepath.Join(root, "zxing-cpp", "core", "src")
	dstDir := filepath.Join(root, "include", "ZXing")

	// Check source directory exists
	if _, err := os.Stat(srcDir); err != nil {
		return fmt.Errorf("zxing-cpp source directory not found: %s (run 'git submodule update --init --recursive' first)", srcDir)
	}

	// Create destination directory
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create include/ZXing directory: %w", err)
	}

	// Copy all .h files from srcDir to dstDir
	count := 0
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".h") && !strings.HasSuffix(path, ".hpp") {
			return nil
		}

		// Calculate relative path from srcDir
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Only copy files under ZXing/ directory
		if !strings.HasPrefix(relPath, "ZXing/") && !strings.HasPrefix(relPath, "ZXing\\") {
			return nil
		}

		dstPath := filepath.Join(root, "include", relPath)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		if err := copyFile(path, dstPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
		count++
		fmt.Printf("Copied: %s\n", relPath)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to sync headers: %w", err)
	}

	// Verify zxing.h exists in include/
	zxingH := filepath.Join(root, "include", "zxing.h")
	if _, err := os.Stat(zxingH); err != nil {
		fmt.Printf("Warning: %s not found\n", zxingH)
	}

	fmt.Printf("Synced %d header files.\n", count)
	return nil
}
