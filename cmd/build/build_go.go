package main

import (
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

// buildAll builds everything: C++ libraries, WASM module, and Go packages.
// Missing build dependencies (CMake, EMSDK) produce a warning and skip that step
// rather than failing the entire build.
func buildAll(args []string) error {
	if err := buildLib(args); err != nil {
		fmt.Printf("Warning: build-lib skipped (%v)\n", err)
	}
	if err := buildWasm(args); err != nil {
		fmt.Printf("Warning: build-wasm skipped (%v)\n", err)
	}
	if err := buildGo(args); err != nil {
		return fmt.Errorf("build-go failed: %w", err)
	}
	fmt.Println("All builds complete.")
	return nil
}
