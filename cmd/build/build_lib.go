package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// buildLib builds C++ static libraries via CMake.
// It runs git submodule update, cmake, and make, then copies artifacts to lib/{os}-{arch}/.
func buildLib(args []string) error {
	// Check build dependencies before starting
	if _, err := exec.LookPath("cmake"); err != nil {
		return errMissingDep{tool: "cmake"}
	}
	makeCmd := "make"
	if runtime.GOOS == "windows" {
		makeCmd = "mingw32-make"
	}
	if _, err := exec.LookPath(makeCmd); err != nil {
		return errMissingDep{tool: makeCmd}
	}

	root, err := projectRoot()
	if err != nil {
		return err
	}

	// Step 1: Update git submodules
	fmt.Println("Updating git submodules...")
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git submodule update failed: %w", err)
	}

	// Step 2: Create build directory
	buildDir := filepath.Join(root, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Step 3: Run CMake
	// Windows uses MinGW Makefiles (not MSVC) because Go CGO on Windows
	// natively uses MinGW toolchain, and MinGW produces .a archives that
	// are compatible with CGO linking. MSVC would produce .lib files
	// that are incompatible with CGO.
	fmt.Println("Running CMake...")
	generator := "Unix Makefiles"
	if runtime.GOOS == "windows" {
		generator = "MinGW Makefiles"
	}

	cmd = exec.Command("cmake", "-G", generator, "-DCMAKE_BUILD_TYPE=Release", root)
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmake failed: %w", err)
	}

	// Step 4: Run make
	fmt.Println("Building...")
	cmd = exec.Command(makeCmd, "-j", fmt.Sprintf("%d", runtime.NumCPU()))
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make failed: %w", err)
	}

	// Step 5: Copy artifacts to lib/{os}-{arch}/
	libPath := filepath.Join(root, libDir())
	if err := os.MkdirAll(libPath, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	// Copy static libraries (CMake + MinGW produces .a on all platforms)
	artifacts := []string{
		filepath.Join(buildDir, "lib", "libZXing.a"),
		filepath.Join(buildDir, "lib", "libzxingwrapper.a"),
	}

	for _, src := range artifacts {
		if _, err := os.Stat(src); err != nil {
			fmt.Printf("Warning: artifact not found: %s\n", src)
			continue
		}
		dst := filepath.Join(libPath, filepath.Base(src))
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", src, err)
		}
		fmt.Printf("Copied: %s -> %s\n", src, dst)
	}

	fmt.Println("Build complete.")
	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
