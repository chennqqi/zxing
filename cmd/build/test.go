package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runTest runs Go tests with CGO environment if available.
// Respects CGO_ENABLED env var: "0" forces non-CGO, "1" forces CGO,
// unset auto-detects based on precompiled library availability.
func runTest(args []string) error {
	env, msg, err := selectBuildEnv()
	if err != nil {
		return err
	}
	fmt.Printf("Test backend: %s\n", msg)

	fmt.Println("Running tests...")
	testArgs := append([]string{"test"}, args...)
	testArgs = append(testArgs, "./pkg/...", "-v")
	cmd := exec.Command("go", testArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	fmt.Println("Tests complete.")
	return nil
}
