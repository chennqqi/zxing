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

	// Check CGO_CXXFLAGS contains -std=c++20
	cxxflags := envGet(env, "CGO_CXXFLAGS")
	if !strings.Contains(cxxflags, "-std=c++20") {
		t.Errorf("CGO_CXXFLAGS should contain -std=c++20, got %s", cxxflags)
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

func TestHasPrebuiltLibs(t *testing.T) {
	// This test runs in the project root where lib/{os}-{arch}/ should exist
	result := hasPrebuiltLibs()
	// We can't guarantee libs exist in all test environments,
	// but the function should not panic and should return a bool
	t.Logf("hasPrebuiltLibs() = %v", result)
}

func TestSelectBuildEnvForcedNonCGO(t *testing.T) {
	t.Setenv("CGO_ENABLED", "0")
	env, msg, err := selectBuildEnv()
	if err != nil {
		t.Fatalf("selectBuildEnv() failed: %v", err)
	}
	if envGet(env, "CGO_ENABLED") != "0" {
		t.Errorf("CGO_ENABLED should be 0, got %s", envGet(env, "CGO_ENABLED"))
	}
	if !strings.Contains(msg, "non-CGO") {
		t.Errorf("msg should mention non-CGO, got %s", msg)
	}
	// Verify no CGO_CFLAGS/CGO_CXXFLAGS/CGO_LDFLAGS leaked
	if envHas(env, "CGO_CFLAGS") {
		t.Error("CGO_CFLAGS should not be present in non-CGO env")
	}
	if envHas(env, "CGO_CXXFLAGS") {
		t.Error("CGO_CXXFLAGS should not be present in non-CGO env")
	}
	if envHas(env, "CGO_LDFLAGS") {
		t.Error("CGO_LDFLAGS should not be present in non-CGO env")
	}
}

func TestSelectBuildEnvForcedCGOWithLibs(t *testing.T) {
	t.Setenv("CGO_ENABLED", "1")
	env, msg, err := selectBuildEnv()
	if err != nil {
		// If libs are missing, error is expected
		if strings.Contains(err.Error(), "precompiled libraries not found") {
			t.Skip("precompiled libs not available in test environment")
		}
		t.Fatalf("selectBuildEnv() failed: %v", err)
	}
	if envGet(env, "CGO_ENABLED") != "1" {
		t.Errorf("CGO_ENABLED should be 1, got %s", envGet(env, "CGO_ENABLED"))
	}
	if !strings.Contains(msg, "CGO") {
		t.Errorf("msg should mention CGO, got %s", msg)
	}
}

func TestSelectBuildEnvForcedCGOWithoutLibs(t *testing.T) {
	t.Setenv("CGO_ENABLED", "1")
	// Temporarily break lib path by setting a non-existent project root
	// We can't easily mock projectRoot, so we test the error path
	// by checking that the function returns an error when libs don't exist
	// This test only works if libs are actually missing
	if hasPrebuiltLibs() {
		t.Skip("precompiled libs exist, cannot test missing-libs error path")
	}
	_, _, err := selectBuildEnv()
	if err == nil {
		t.Fatal("selectBuildEnv() should return error when CGO_ENABLED=1 but libs missing")
	}
	if !strings.Contains(err.Error(), "precompiled libraries not found") {
		t.Errorf("error should mention missing libs, got: %v", err)
	}
}

func TestSelectBuildEnvAutoDetect(t *testing.T) {
	// Unset CGO_ENABLED — t.Setenv with empty string simulates unset
	t.Setenv("CGO_ENABLED", "")
	env, msg, err := selectBuildEnv()
	if err != nil {
		t.Fatalf("selectBuildEnv() failed: %v", err)
	}
	// Should auto-detect based on lib availability
	if hasPrebuiltLibs() {
		if envGet(env, "CGO_ENABLED") != "1" {
			t.Errorf("auto-detect should use CGO when libs exist, got CGO_ENABLED=%s", envGet(env, "CGO_ENABLED"))
		}
		if !strings.Contains(msg, "auto") {
			t.Errorf("msg should mention auto-detect, got %s", msg)
		}
	} else {
		if envGet(env, "CGO_ENABLED") != "0" {
			t.Errorf("auto-detect should use non-CGO when libs missing, got CGO_ENABLED=%s", envGet(env, "CGO_ENABLED"))
		}
		if !strings.Contains(msg, "non-CGO") {
			t.Errorf("msg should mention non-CGO, got %s", msg)
		}
	}
}

func TestSelectBuildEnvNonStandardValue(t *testing.T) {
	t.Setenv("CGO_ENABLED", "true")
	_, msg, err := selectBuildEnv()
	if err != nil {
		t.Fatalf("selectBuildEnv() failed for non-standard value: %v", err)
	}
	// Non-standard values should be treated as auto-detect
	if !strings.Contains(msg, "auto") && !strings.Contains(msg, "non-CGO") {
		t.Errorf("non-standard CGO_ENABLED should auto-detect, got msg: %s", msg)
	}
}

func TestBuildNonCGOEnvCleansAllCGOVars(t *testing.T) {
	// Set CGO_* vars in the process environment
	t.Setenv("CGO_CFLAGS", "-I/fake")
	t.Setenv("CGO_CXXFLAGS", "-std=c++20")
	t.Setenv("CGO_LDFLAGS", "-L/fake -lfoo")

	env := buildNonCGOEnv()
	if envHas(env, "CGO_CFLAGS") {
		t.Error("CGO_CFLAGS should be removed in non-CGO env")
	}
	if envHas(env, "CGO_CXXFLAGS") {
		t.Error("CGO_CXXFLAGS should be removed in non-CGO env")
	}
	if envHas(env, "CGO_LDFLAGS") {
		t.Error("CGO_LDFLAGS should be removed in non-CGO env")
	}
	if envGet(env, "CGO_ENABLED") != "0" {
		t.Errorf("CGO_ENABLED should be 0, got %s", envGet(env, "CGO_ENABLED"))
	}
}
