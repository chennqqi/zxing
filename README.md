# ZXing Go

English | [中文](README_CN.md)

[![Release](https://github.com/chennqqi/zxing/actions/workflows/release.yml/badge.svg)](https://github.com/chennqqi/zxing/actions/workflows/release.yml)

A Go barcode recognition library based on [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp), supporting both CGO and WASM (wazero) backends.

## Features

- Supports multiple barcode formats: QR Code, Aztec, Codabar, Code 39/93/128, Data Matrix, EAN-8/13, ITF, MaxiCode, PDF417, UPC-A/E
- Supports single and multiple barcode decoding
- Dual-backend architecture:
  - **CGO backend**: native C++ static library, highest performance
  - **WASM backend**: pure Go WASM runtime based on [wazero](https://github.com/tetratelabs/wazero), no CGO dependency
- Compile-time backend selection via Go build tags
- Cross-platform support: Linux, Windows, macOS
- Prebuilt static libraries: bundled for Linux / Windows / macOS (x64 and arm64), no local C++ toolchain required

## Platform Support

| Platform | CGO Backend | WASM Backend | Notes |
|------|----------|----------|------|
| Linux x64 | ✅ | ✅ | Prebuilt static library included |
| Linux arm64 | ✅ | ✅ | Prebuilt static library (glibc 2.28+) |
| Windows x64 | ✅ | ✅ | Prebuilt static library included (requires MinGW-w64 GCC) |
| Windows arm64 | ✅ | ✅ | Prebuilt static library (requires clang / llvm-mingw toolchain) |
| macOS x64 / arm64 | ✅ | ✅ | Prebuilt static library (requires Xcode clang) |
| js/wasm | ❌ | ✅ | Browser/Node.js environments |

> **Windows CGO notes**: The CGO backend on Windows x64 requires the **MinGW-w64** GCC toolchain (not MSVC).
> The prebuilt static libraries are compiled with MinGW-w64 and are compatible with Go CGO's default toolchain.
> Install via `choco install mingw` or download from [winlibs.com](https://winlibs.com/).
>
> **Windows arm64**: MinGW-w64 GCC has no aarch64 target; CGO requires the **clang** toolchain
> (llvm-mingw or MSYS2 CLANGARM64). Set `CC=clang CXX=clang++` when building.

## Backend Selection

The backend is selected automatically at compile time via build tags:

| Condition | Backend | Notes |
|------|------|------|
| `CGO_ENABLED=1` + Linux/Windows/macOS | CGO | Native C++ static library |
| `CGO_ENABLED=0` or other platforms | WASM (wazero) | Pure Go WASM runtime |
| `GOOS=js GOARCH=wasm` | WASM (js) | Browser/Node.js environments |

You can also specify it manually via `Config.Backend`:

```go
// Automatic selection (default)
zx, _ := zxing.New(nil)

// Force WASM
zx, _ := zxing.New(&zxing.Config{Backend: zxing.BackendWASM})

// Force CGO
zx, _ := zxing.New(&zxing.Config{Backend: zxing.BackendCGO})
```

## Quick Start

### WASM Backend (no CGO required)

```bash
# Clone the project
git clone --recursive https://github.com/chennqqi/zxing.git
cd zxing

# Build with the WASM backend (default)
CGO_ENABLED=0 go build ./pkg/zxing/

# Run tests
CGO_ENABLED=0 go test ./pkg/zxing/ -v
```

### CGO Backend

```bash
# Using the build tool (recommended)
go run ./cmd/build build-go

# Or build manually
CGO_ENABLED=1 go build ./pkg/zxing/
```

## Usage Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    // Automatically select a backend
    zx, err := zxing.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer zx.Close()

    // Decode an image
    result, err := zx.DecodeImage(context.Background(), img, &zxing.DecodeOptions{
        TryHarder: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Decoded: %s (format: %s, backend: %s)\n",
        result.Text, result.Format, zx.GetBackend())
}
```

## Build Tool

The project provides a unified build tool in `cmd/build/`, replacing all fragmented scripts:

```bash
# Build the Go package
go run ./cmd/build build-go

# Build the C++ static libraries
go run ./cmd/build build-lib

# Build the WASM module (requires Emscripten)
go run ./cmd/build build-wasm

# Build everything
go run ./cmd/build build-all

# Sync ZXing-CPP headers
go run ./cmd/build sync-headers

# Run tests
go run ./cmd/build test

# Clean build artifacts
go run ./cmd/build clean

# Build Linux static libraries in Docker (CentOS 7, glibc 2.17, GCC 10, C++20)
go run ./cmd/build docker-build
```

### Rebuilding the Static Libraries

If you need to rebuild the C++ static libraries (e.g. after updating the zxing-cpp submodule):

```bash
# Linux x64 (via Docker, CentOS 7 + devtoolset-10, glibc 2.17 compatible)
docker build -t zxing-linux-build -f docker/Dockerfile.linux-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-linux-build \
  sh -c "/tmp/patch_using_enum.sh /workspace/zxing-cpp && \
  cd /tmp && rm -rf build && mkdir -p build && cd build && \
  cmake3 -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF \
  -DCMAKE_CXX_STANDARD=20 -DCMAKE_CXX_FLAGS=-fcoroutines /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/"

# Windows x64 (via Docker cross-compilation, MinGW-w64)
docker build -t zxing-win-build -f docker/Dockerfile.win-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-win-build \
  sh -c "cd /tmp && mkdir -p build && cd build && \
  cmake -DCMAKE_TOOLCHAIN_FILE=/workspace/docker/mingw-w64-toolchain.cmake \
  -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/windows-x64/"
```

Static libraries for the other platforms (linux-arm64, windows-arm64, darwin-x64, darwin-arm64) are built via
GitHub Actions: manually trigger the **Build Prebuilt Libraries** workflow on the repository's Actions page.
The artifacts are uploaded as workflow artifacts (existing libraries are never overwritten); download them and
commit them to the `lib/{os}-{arch}/` directory.

> **Linux compatibility**: The Linux static libraries are compiled on CentOS 7 (glibc 2.17) and are compatible with CentOS 7+, Ubuntu 16.04+, RHEL 7+, Debian 8+, and all other distributions using glibc 2.17+.
> GCC 10 requires the patch script (`docker/patch_using_enum.sh`) to replace `using enum` syntax, plus the `-fcoroutines` flag.

### Building a CGO Binary Compatible with Older Systems

If you need the CGO binary to also run on older systems (glibc 2.17+), **the entire build must run inside the CentOS 7 container** (including `go build`):

```bash
# Build the C++ static libraries + Go CGO binary entirely inside the CentOS 7 container
docker build -t zxing-linux-build -f docker/Dockerfile.linux-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-linux-build sh -c "
  /tmp/patch_using_enum.sh /workspace/zxing-cpp && \
  cd /tmp && rm -rf build && mkdir -p build && cd build && \
  cmake3 -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF \
  -DCMAKE_CXX_STANDARD=20 -DCMAKE_CXX_FLAGS=-fcoroutines /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/ && \
  cd /workspace && \
  CGO_ENABLED=1 \
  CGO_CFLAGS='-I/workspace/include' \
  CGO_CXXFLAGS='-std=c++20 -I/workspace/include' \
  CGO_LDFLAGS='-L/workspace/lib/linux-x64 -lzxingwrapper -lZXing -lstdc++ -lm' \
  go build -buildvcs=false -o zxing-cli ./cmd/zxing-cli
"
```

> **Why must `go build` also run inside the container?** Static libraries (.a) carry no glibc version markers themselves, but when `go build` links the final binary, the linker resolves symbol versions against the host system's glibc. Linking on a host with glibc 2.39 produces a binary requiring GLIBC_2.38, even if the static libraries were built on CentOS 7. Building everything inside CentOS 7 caps the highest GLIBC requirement at 2.14.

## Environment Variables

```bash
# Specify the backend
export ZXING_BACKEND=auto    # auto | cgo | wasm

# Specify the WASM module path
export ZXING_WASM_PATH=wasm/zxingwrapper.wasm

# Enable debug mode
export ZXING_DEBUG=true

# Timeout (seconds)
export ZXING_TIMEOUT=30
```

## API

### ZXing Interface

```go
type ZXing interface {
    DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error)
    DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error)
    EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error)
    EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error)
    Close() error
    GetBackend() Backend
}
```

### Factory Methods

```go
// Automatically select a backend
zx, err := zxing.New(nil)

// Use the CGO backend
zx, err := zxing.NewCGO(config)

// Use the WASM backend
zx, err := zxing.NewWASM(config)
```

## Project Structure

```
zxing/
├── cmd/
│   ├── build/                 # Unified build tool
│   ├── zxing-cli/             # Command-line tool
│   └── server/                # HTTP server
├── pkg/
│   ├── zxing/                 # Unified interface layer
│   │   ├── interface.go       # ZXing interface definition
│   │   ├── config.go          # Configuration
│   │   ├── factory.go         # Backend factory
│   │   ├── cgo_binding.go     # CGO bindings (linux/windows/darwin × amd64/arm64)
│   │   ├── cgo_impl.go        # CGO implementation
│   │   ├── cgo_stub.go        # CGO stub (non-CGO platforms)
│   │   ├── wasm_impl.go       # wazero WASM implementation
│   │   ├── wasm_impl_js.go    # js/wasm implementation
│   │   └── wasm_stub.go       # WASM stub (CGO platforms)
│   └── wasm/                  # WASM runtime
│       ├── runtime_wazero.go  # wazero runtime
│       ├── runtime_js.go      # js/wasm runtime
│       └── runtime_stub.go    # Runtime stub
├── include/                   # C/C++ headers
├── lib/                       # Prebuilt static libraries
├── wasm/                      # WASM module
├── docker/                    # Docker build environments
├── src/                       # C++ wrapper source
└── CMakeLists.txt             # CMake build configuration
```

## Performance Comparison

| Backend | Compile Time | Runtime Performance | Memory Usage | Portability | CGO Dependency |
|------|----------|----------|----------|----------|----------|
| CGO  | Slow     | Fastest  | Low      | Limited  | Yes      |
| WASM | Fast     | Fast     | Medium   | Excellent| No       |

## Testing

```bash
# WASM backend tests
CGO_ENABLED=0 go test ./pkg/zxing/ -v

# CGO backend tests
CGO_ENABLED=1 go test ./pkg/zxing/ -v

# Build tool tests
go test ./cmd/build/ -v

# WASM runtime tests
CGO_ENABLED=0 go test ./pkg/wasm/ -v
```

## License

MIT License

## Related Links

- [ZXing official project](https://github.com/zxing/zxing)
- [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp)
- [wazero](https://github.com/tetratelabs/wazero)
- [WebAssembly](https://webassembly.org/)
