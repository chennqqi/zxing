package main

import (
	"fmt"
	"os"
	"os/exec"
)

// buildGo builds Go packages with CGO environment if available.
func buildGo(args []string) error {
	env, err := buildCGOEnv()
	if err != nil {
		fmt.Printf("Warning: CGO env setup failed (%v), using non-CGO build\n", err)
		env = buildNonCGOEnv()
	}

	fmt.Println("Building Go packages...")
	cmd := exec.Command("go", "build", "./...")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// Also build the CLI
	fmt.Println("Building zxing-cli...")
	cmd = exec.Command("go", "build", "-o", "bin/zxing-cli", "./cmd/zxing-cli")
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
func buildAll(args []string) error {
	if err := buildLib(args); err != nil {
		return fmt.Errorf("build-lib failed: %w", err)
	}
	if err := buildWasm(args); err != nil {
		return fmt.Errorf("build-wasm failed: %w", err)
	}
	if err := buildGo(args); err != nil {
		return fmt.Errorf("build-go failed: %w", err)
	}
	fmt.Println("All builds complete.")
	return nil
}
