package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// dockerBuild builds Linux static library in a CentOS 7 Docker container.
// This ensures glibc 2.17 compatibility for the precompiled library.
// The container includes devtoolset-10 (GCC 10), cmake3, and Go 1.24.
// zxing-cpp v3.0.2 requires C++20 features patched by patch_using_enum.sh.
func dockerBuild(args []string) error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(root, "docker", "Dockerfile.linux-build")

	// Check if Dockerfile exists
	if _, err := os.Stat(dockerfilePath); err != nil {
		return fmt.Errorf("Dockerfile not found: %s", dockerfilePath)
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
		"-v", fmt.Sprintf("%s:/workspace:Z", root),
		"zxing-linux-build",
		"bash", "-c",
		"/tmp/patch_using_enum.sh /workspace/zxing-cpp && "+
			"cd /tmp && rm -rf build && mkdir -p build && cd build && "+
			"cmake3 -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF "+
			"-DCMAKE_CXX_STANDARD=20 -DCMAKE_CXX_FLAGS=-fcoroutines /workspace && "+
			"make -j$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run failed: %w", err)
	}

	fmt.Println("Docker build complete.")
	return nil
}
