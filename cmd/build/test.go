package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runTest runs Go tests with CGO environment if available.
func runTest(args []string) error {
	env, err := buildCGOEnv()
	if err != nil {
		fmt.Printf("Warning: CGO env setup failed (%v), using non-CGO build\n", err)
		env = buildNonCGOEnv()
	}

	fmt.Println("Running tests...")
	cmd := exec.Command("go", "test", "./pkg/...", "-v")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	fmt.Println("Tests complete.")
	return nil
}
