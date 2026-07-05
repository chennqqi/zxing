package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// detectArch returns the architecture suffix for library paths.
// Returns "x64" for amd64 and "arm64" for arm64.
func detectArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64"
	default:
		return "x64"
	}
}

// libDir returns the platform-specific library directory path.
// Format: lib/{os}-{arch}
func libDir() string {
	return fmt.Sprintf("lib/%s-%s", runtime.GOOS, detectArch())
}

// absPath converts a relative path to an absolute path based on the project root.
func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	// Find project root by looking for go.mod
	root, err := projectRoot()
	if err != nil {
		// Fallback: use current directory
		wd, _ := os.Getwd()
		return filepath.Join(wd, p)
	}
	return filepath.Join(root, p)
}

// projectRoot returns the absolute path to the project root (where go.mod is).
func projectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// buildCGOEnv returns environment variables for CGO-enabled builds.
// It starts from os.Environ() and overrides CGO-related variables.
func buildCGOEnv() ([]string, error) {
	root, err := projectRoot()
	if err != nil {
		return nil, err
	}

	includeDir := filepath.Join(root, "include")
	libPath := filepath.Join(root, libDir())

	cflags := fmt.Sprintf("-I%s", includeDir)
	cxxflags := fmt.Sprintf("-std=c++20 -I%s", includeDir)
	ldflags := fmt.Sprintf("-L%s -lzxingwrapper -lZXing -lstdc++", libPath)
	if runtime.GOOS == "linux" {
		ldflags += " -lm"
	}

	env := make([]string, 0, len(os.Environ())+4)
	for _, e := range os.Environ() {
		// Skip existing CGO_ variables to avoid duplicates
		if strings.HasPrefix(e, "CGO_ENABLED=") ||
			strings.HasPrefix(e, "CGO_CFLAGS=") ||
			strings.HasPrefix(e, "CGO_CXXFLAGS=") ||
			strings.HasPrefix(e, "CGO_LDFLAGS=") {
			continue
		}
		env = append(env, e)
	}

	// Add CGO-related variables
	env = append(env,
		"CGO_ENABLED=1",
		fmt.Sprintf("CGO_CFLAGS=%s", cflags),
		fmt.Sprintf("CGO_CXXFLAGS=%s", cxxflags),
		fmt.Sprintf("CGO_LDFLAGS=%s", ldflags),
	)

	return env, nil
}

// buildNonCGOEnv returns environment variables for non-CGO builds.
func buildNonCGOEnv() []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "CGO_ENABLED=") {
			env = append(env, e)
		}
	}
	env = append(env, "CGO_ENABLED=0")
	return env
}

// envHas checks if an environment variable exists in the env slice.
func envHas(env []string, key string) bool {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return true
		}
	}
	return false
}

// envGet returns the value of an environment variable from the env slice.
func envGet(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return e[len(prefix):]
		}
	}
	return ""
}
