package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// buildGo builds Go packages with CGO environment if available.
// Respects CGO_ENABLED env var: "0" forces non-CGO, "1" forces CGO,
// unset auto-detects based on precompiled library availability.
func buildGo(args []string) error {
	env, msg, err := selectBuildEnv()
	if err != nil {
		return err
	}
	fmt.Printf("Build backend: %s\n", msg)

	fmt.Println("Building Go packages...")
	buildArgs := append([]string{"build"}, args...)
	buildArgs = append(buildArgs, "./...")
	cmd := exec.Command("go", buildArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// Also build the CLI as a deliverable
	fmt.Println("Building zxing-cli...")
	outputPath := "bin/zxing-cli"
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	cmd = exec.Command("go", "build", "-o", outputPath, "./cmd/zxing-cli")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build zxing-cli failed: %w", err)
	}

	fmt.Println("Go build complete.")
	return nil
}

// errMissingDep indicates a required build dependency is not installed.
type errMissingDep struct {
	tool string
}

func (e errMissingDep) Error() string {
	return fmt.Sprintf("missing build dependency: %s", e.tool)
}

// isDepMissingError returns true if the error indicates a missing build
// dependency (e.g. cmake or emcmake command not found, EMSDK not set).
// Uses errors.Is and errors.As for reliable cross-platform detection.
func isDepMissingError(err error) bool {
	if err == nil {
		return false
	}
	var md errMissingDep
	return errors.As(err, &md) || errors.Is(err, exec.ErrNotFound)
}

// buildAll builds everything: C++ libraries, WASM module, and Go packages.
// Steps with missing build dependencies (CMake, EMSDK) are skipped with a
// warning; source compilation failures return a fatal error.
func buildAll(args []string) error {
	if err := buildLib(args); err != nil {
		if isDepMissingError(err) {
			fmt.Printf("Warning: build-lib skipped (dependency missing: %v)\n", err)
		} else {
			return fmt.Errorf("build-lib failed: %w", err)
		}
	}
	if err := buildWasm(args); err != nil {
		if isDepMissingError(err) {
			fmt.Printf("Warning: build-wasm skipped (dependency missing: %v)\n", err)
		} else {
			return fmt.Errorf("build-wasm failed: %w", err)
		}
	}
	if err := buildGo(args); err != nil {
		return fmt.Errorf("build-go failed: %w", err)
	}
	fmt.Println("All builds complete.")
	return nil
}
