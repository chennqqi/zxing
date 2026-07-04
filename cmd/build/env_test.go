package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCGOEnv(t *testing.T) {
	env, err := buildCGOEnv()
	if err != nil {
		t.Fatalf("buildCGOEnv() failed: %v", err)
	}

	// Check CGO_ENABLED=1 is present
	if !envHas(env, "CGO_ENABLED") {
		t.Fatal("CGO_ENABLED not found in env")
	}
	if envGet(env, "CGO_ENABLED") != "1" {
		t.Errorf("CGO_ENABLED should be 1, got %s", envGet(env, "CGO_ENABLED"))
	}

	// Check CGO_CFLAGS contains -I
	cflags := envGet(env, "CGO_CFLAGS")
	if !strings.Contains(cflags, "-I") {
		t.Errorf("CGO_CFLAGS should contain -I, got %s", cflags)
	}

	// Check CGO_CXXFLAGS contains -std=c++17
	cxxflags := envGet(env, "CGO_CXXFLAGS")
	if !strings.Contains(cxxflags, "-std=c++17") {
		t.Errorf("CGO_CXXFLAGS should contain -std=c++17, got %s", cxxflags)
	}

	// Check CGO_LDFLAGS contains -lzxingwrapper
	ldflags := envGet(env, "CGO_LDFLAGS")
	if !strings.Contains(ldflags, "-lzxingwrapper") {
		t.Errorf("CGO_LDFLAGS should contain -lzxingwrapper, got %s", ldflags)
	}
}

func TestBuildCGOEnvDoesNotModifyOSEnviron(t *testing.T) {
	original := os.Environ()
	_, _ = buildCGOEnv()
	after := os.Environ()

	if len(original) != len(after) {
		t.Errorf("os.Environ() was modified: original len=%d, after len=%d", len(original), len(after))
	}
}

func TestDetectArch(t *testing.T) {
	arch := detectArch()
	if arch != "x64" && arch != "arm64" {
		t.Errorf("detectArch() should return x64 or arm64, got %s", arch)
	}
}

func TestLibDir(t *testing.T) {
	dir := libDir()
	if !strings.Contains(dir, "lib/") {
		t.Errorf("libDir() should contain 'lib/', got %s", dir)
	}
}

func TestAbsPath(t *testing.T) {
	p := absPath("include")
	if !filepath.IsAbs(p) {
		t.Errorf("absPath() should return absolute path, got %s", p)
	}
}
