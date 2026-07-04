package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// dockerBuild builds Linux static library in a CentOS 7 Docker container.
// This ensures glibc 2.17 compatibility for the precompiled library.
func dockerBuild(args []string) error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(root, "docker", "Dockerfile.linux-build")

	// Check if Dockerfile exists
	if _, err := os.Stat(dockerfilePath); err != nil {
		return fmt.Errorf("Dockerfile not found: %s (run build tool first to create it)", dockerfilePath)
	}

	// Build Docker image
	fmt.Println("Building Docker image...")
	cmd := exec.Command("docker", "build", "-t", "zxing-linux-build", "-f", dockerfilePath, filepath.Join(root, "docker"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// Run container to build the library
	fmt.Println("Running build in Docker container...")
	libDir := filepath.Join(root, "lib", "linux-x64")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	cmd = exec.Command("docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", root),
		"-v", fmt.Sprintf("%s:/workspace/lib/linux-x64", libDir),
		"zxing-linux-build",
		"bash", "-c",
		"cd /workspace && cmake -G 'Unix Makefiles' -DCMAKE_BUILD_TYPE=Release . && make -j$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run failed: %w", err)
	}

	fmt.Println("Docker build complete.")
	return nil
}
