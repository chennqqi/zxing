package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// detectImageType detects image type from magic number
func detectImageType(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && bytes.Equal(data[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ".png"
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && bytes.Equal(data[0:3], []byte{0xFF, 0xD8, 0xFF}) {
		return ".jpg"
	}

	// GIF: 47 49 46 38 (GIF8) or 47 49 46 39 (GIF9)
	if len(data) >= 4 {
		if bytes.Equal(data[0:4], []byte{0x47, 0x49, 0x46, 0x38}) || 
		   bytes.Equal(data[0:4], []byte{0x47, 0x49, 0x46, 0x39}) {
			return ".gif"
		}
	}

	// BMP: 42 4D (BM)
	if len(data) >= 2 && bytes.Equal(data[0:2], []byte{0x42, 0x4D}) {
		return ".bmp"
	}

	// WebP: RIFF...WEBP
	if len(data) >= 12 {
		if bytes.Equal(data[0:4], []byte{0x52, 0x49, 0x46, 0x46}) && // RIFF
		   bytes.Equal(data[8:12], []byte{0x57, 0x45, 0x42, 0x50}) { // WEBP
			return ".webp"
		}
	}

	return ""
}

func main() {
	imagesDir := "data/images"
	if len(os.Args) > 1 {
		imagesDir = os.Args[1]
	}

	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", imagesDir, err)
		os.Exit(1)
	}

	renamed := 0
	skipped := 0
	errors := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		oldPath := filepath.Join(imagesDir, entry.Name())
		
		// Skip files that already have extensions
		if strings.Contains(entry.Name(), ".") && 
		   !strings.HasPrefix(entry.Name(), ".") {
			skipped++
			continue
		}

		// Read first 16 bytes to detect magic number
		file, err := os.Open(oldPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", oldPath, err)
			errors++
			continue
		}

		header := make([]byte, 16)
		n, err := file.Read(header)
		file.Close()
		
		if err != nil || n < 4 {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", oldPath, err)
			errors++
			continue
		}

		ext := detectImageType(header)
		if ext == "" {
			skipped++
			continue
		}

		newPath := oldPath + ext
		
		// Check if target file already exists
		if _, err := os.Stat(newPath); err == nil {
			fmt.Printf("Skipping %s: target %s already exists\n", entry.Name(), newPath)
			skipped++
			continue
		}

		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error renaming %s to %s: %v\n", oldPath, newPath, err)
			errors++
			continue
		}

		fmt.Printf("Renamed: %s -> %s\n", entry.Name(), entry.Name()+ext)
		renamed++
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Renamed: %d files\n", renamed)
	fmt.Printf("  Skipped: %d files\n", skipped)
	fmt.Printf("  Errors: %d files\n", errors)
}
