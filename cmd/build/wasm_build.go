package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// buildWasm builds the WASM module via Emscripten.
// It checks for EMSDK, backs up CMakeLists.txt, uses CMakeLists-wasm.txt,
// runs emcmake + emmake, copies artifacts to wasm/, and restores CMakeLists.txt.
func buildWasm(args []string) error {
	// Check build dependencies before starting
	if os.Getenv("EMSDK") == "" {
		return errMissingDep{tool: "EMSDK environment variable"}
	}
	if _, err := exec.LookPath("emcmake"); err != nil {
		return errMissingDep{tool: "emcmake"}
	}
	if _, err := exec.LookPath("emmake"); err != nil {
		return errMissingDep{tool: "emmake"}
	}

	root, err := projectRoot()
	if err != nil {
		return err
	}

	// Backup CMakeLists.txt
	cmakeFile := filepath.Join(root, "CMakeLists.txt")
	cmakeWasmFile := filepath.Join(root, "CMakeLists-wasm.txt")
	backupFile := filepath.Join(root, "CMakeLists.txt.bak")

	if err := copyFile(cmakeFile, backupFile); err != nil {
		return fmt.Errorf("failed to backup CMakeLists.txt: %w", err)
	}
	defer func() {
		_ = os.Rename(backupFile, cmakeFile)
		fmt.Println("Restored CMakeLists.txt")
	}()

	// Copy CMakeLists-wasm.txt as CMakeLists.txt
	if err := copyFile(cmakeWasmFile, cmakeFile); err != nil {
		return fmt.Errorf("failed to copy CMakeLists-wasm.txt: %w", err)
	}

	// Create build-wasm directory
	buildDir := filepath.Join(root, "build-wasm")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build-wasm directory: %w", err)
	}

	// Run emcmake cmake
	fmt.Println("Running emcmake cmake...")
	cmd := exec.Command("emcmake", "cmake", root, "-G", "Unix Makefiles", "-DCMAKE_BUILD_TYPE=Release")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("emcmake cmake failed: %w", err)
	}

	// Run emmake make
	fmt.Println("Running emmake make...")
	cmd = exec.Command("emmake", "make", "-j", fmt.Sprintf("%d", numCPU()))
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("emmake make failed: %w", err)
	}

	// Copy artifacts to wasm/
	wasmDir := filepath.Join(root, "wasm")
	if err := os.MkdirAll(wasmDir, 0755); err != nil {
		return fmt.Errorf("failed to create wasm directory: %w", err)
	}

	wasmSrc := filepath.Join(buildDir, "bin", "zxingwrapper.wasm")
	wasmDst := filepath.Join(wasmDir, "zxingwrapper.wasm")
	if err := copyFile(wasmSrc, wasmDst); err != nil {
		return fmt.Errorf("failed to copy WASM artifact: %w", err)
	}
	fmt.Printf("Copied: %s -> %s\n", wasmSrc, wasmDst)

	fmt.Println("WASM build complete.")
	return nil
}

// numCPU returns the number of CPU cores.
func numCPU() int {
	return runtime.NumCPU()
}
