package main

import (
	"fmt"
	"os"
)

const usageText = `zxing build tool — cross-platform build utility

Usage: go run ./cmd/build <command> [args]

Commands:
  build-lib      Build C++ static libraries via CMake
  build-wasm     Build WASM module via Emscripten
  build-go       Build Go packages (auto-detects CGO or non-CGO)
  build-all      Build everything (lib + wasm + go, skips missing deps).
                 Additional args are passed to the final go build step.
  sync-headers   Sync ZXing-CPP headers to include/ZXing/
  test           Run Go tests (auto-detects CGO or non-CGO)
  clean          Remove build artifacts (build/ and build-wasm/)
  docker-build   Build Linux static library in CentOS 7 Docker container

Environment:
  CGO_ENABLED=0  Force non-CGO (WASM) backend
  CGO_ENABLED=1  Force CGO backend (requires precompiled libraries)
  (unset/other)  Auto-detect: CGO if libs exist, otherwise non-CGO

Examples:
  go run ./cmd/build build-go
  go run ./cmd/build build-go -v
  CGO_ENABLED=0 go run ./cmd/build test
  go run ./cmd/build build-all
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usageText)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "build-lib":
		err = buildLib(args)
	case "build-wasm":
		err = buildWasm(args)
	case "build-go":
		err = buildGo(args)
	case "build-all":
		err = buildAll(args)
	case "sync-headers":
		err = syncHeaders(args)
	case "test":
		err = runTest(args)
	case "clean":
		err = clean(args)
	case "docker-build":
		err = dockerBuild(args)
	case "help", "-h", "--help":
		fmt.Print(usageText)
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		fmt.Print(usageText)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
