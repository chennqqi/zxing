# Production-Ready Design: zxing Go Library

## Overview

Refactor the zxing Go wrapper project to be production-ready with stable, easy-to-use builds across Windows and Linux, supporting both CGO link and WASM (no CGO) backends.

## Problem Statement

- Windows and Linux builds were maintained independently, causing one side to break when the other was modified
- Build scripts are fragmented across many sh/ps1/bat files with overlapping, inconsistent logic
- Direct `go build` fails because CGO_CFLAGS/CGO_LDFLAGS are not configured
- WASM backend only works in browser environment (syscall/js), not usable in server-side Go
- Pre-compiled libraries are scattered across `lib/`, `lib64/`, `wasm/` without unified structure

## Solution: Pre-compiled Libraries + wazero WASM Fallback

### Key Decisions

1. **Pre-compiled static libraries committed to repo** — Linux built in Docker CentOS 7 (glibc 2.17) for maximum compatibility; Windows built with MSVC
2. **wazero pure-Go WASM runtime** — WASM backend works in server-side Go via wazero, not just browsers
3. **Go-based cross-platform build tool** (`cmd/build/`) — replaces all fragmented shell/PowerShell scripts
4. **Build tags determine backend at compile time** — no runtime backend switching

## Architecture

### Package Structure

```
pkg/zxing/
  interface.go              # ZXing interface, Result, Options type definitions
  config.go                 # Config, Backend enum
  factory.go                # New() / NewCGO() / NewWASM() factory
  cgo_impl.go               # //go:build cgo — CGO backend implementation (platform-agnostic Go code)
  cgo_binding.go            # //go:build cgo && linux — Linux CGO C bindings (#cgo directives + type constants)
  cgo_binding_windows.go    # //go:build cgo && windows — Windows CGO C bindings
  cgo_stub.go               # //go:build !cgo — CGO unavailable stub
  wasm_impl.go              # //go:build !cgo — wazero WASM backend implementation (server-side)
  wasm_impl_js.go           # //go:build js && wasm — browser WASM implementation (retained)
  wasm_stub.go              # //go:build cgo && !(js && wasm) — WASM stub when CGO is active
```

```
pkg/wasm/
  runtime.go                # //go:build !cgo — wazero runtime implementation
  runtime_js.go             # //go:build js && wasm — browser syscall/js runtime (retained)
  runtime_stub.go           # //go:build cgo && !(js && wasm) — stub
```

### Backend Selection Matrix

| Environment | CGO Backend | WASM Backend | Default Backend |
|-------------|-------------|--------------|-----------------|
| CGO_ENABLED=1 (Linux/Windows) | Available | Stub | CGO |
| CGO_ENABLED=0 (Linux/Windows) | Stub | Available (wazero) | WASM |
| GOOS=js GOARCH=wasm | Stub | Available (syscall/js) | WASM |

### Factory Logic

`factory.go` `New(config)` with `BackendAuto`:
- Compile-time: only one backend is available (enforced by build tags)
- `BackendAuto` returns the single available backend
- `BackendCGO`/`BackendWASM` explicit request returns error if that backend is a stub

### Simplification of universal_impl.go

Current `universal_impl.go` contains runtime backend switching logic (CGO fallback to WASM). After refactor:
- `//go:build cgo && !(js && wasm)`: only CGO implementation, calls `cgoZXing` directly
- `//go:build !cgo`: only WASM implementation, calls `wasmZXing` directly
- Runtime switching logic deleted; compile-time build tags ensure only one backend is available

## CGO Backend Design

### Platform-specific CGO bindings

`cgo_binding.go` (`//go:build cgo && linux`):
```go
#cgo CXXFLAGS: -std=c++17
#cgo CFLAGS: -I${SRCDIR}/../../include
#cgo LDFLAGS: -L${SRCDIR}/../../lib/linux-x64 -lzxingwrapper -lZXing -lstdc++ -lm
```

`cgo_binding_windows.go` (`//go:build cgo && windows`):
```go
#cgo CXXFLAGS: -std=c++17
#cgo CFLAGS: -I${SRCDIR}/../../include
#cgo LDFLAGS: -L${SRCDIR}/../../lib/windows-x64 -lzxingwrapper -lZXing
```

Both files share the same Go code (type constants, `Decode`, `DecodeMulti` functions) — only `#cgo` directives differ. Shared Go code lives in `cgo_impl.go` (`//go:build cgo`).

### Pre-compiled Library Sources

| Platform | Build Environment | Artifacts | Compatibility |
|----------|-------------------|----------|---------------|
| Linux x64 | Docker CentOS 7 (glibc 2.17, devtoolset-7 GCC 7) | `libZXing.a`, `libzxingwrapper.a` | RHEL 7-10, Debian 9+, Ubuntu 16+ |
| Windows x64 | Windows + MSVC 2022 | `ZXing.lib`, `zxingwrapper.lib` | Windows 10+ |

### Header File Management

`include/` directory contains:
- `zxing.h`, `zxing_internal.h` — project's own C wrapper headers (already in repo)
- `ZXing/` — public headers copied from `zxing-cpp/core/src/`

`cmd/build sync-headers` subcommand syncs headers from the zxing-cpp submodule after build.

## WASM Backend Design (wazero)

### wazero Integration

`go.mod` adds:
```
github.com/tetratelabs/wazero v1.8.0
```

### Runtime Structure

```go
type Runtime struct {
    module   api.Module
    ctx      context.Context
    decodeFn api.Function
    encodeFn api.Function
}
```

### Initialization Flow

1. Read `wasm/zxingwrapper.wasm` file
2. Compile and instantiate WASM module with wazero
3. Export required host functions (memory allocation)
4. Obtain references to exported functions (`decode_barcode`, `encode_text`, etc.)

### WASM File Source

`wasm/zxingwrapper.wasm` is built from source via EMSDK + `CMakeLists-wasm.txt`, committed to repo. `cmd/build build-wasm` subcommand rebuilds it.

## Build Tool: cmd/build/

### Command Structure

```
go run ./cmd/build/ <subcommand> [flags]
```

| Subcommand | Function | Description |
|------------|----------|-------------|
| `build-lib` | Compile C++ static library from source | Calls CMake, outputs to `lib/` |
| `build-wasm` | Compile WASM from source | Calls EMSDK + CMake, outputs to `wasm/` |
| `build-go` | Compile Go library and CLI | Auto-detects pre-compiled libs, sets CGO env vars |
| `build-all` | build-lib + build-wasm + build-go | One-click full build |
| `sync-headers` | Sync headers to `include/` | Copies public headers from zxing-cpp submodule |
| `test` | Run tests | Auto-sets CGO env vars, runs `go test` |
| `clean` | Clean build artifacts | Removes `build/`, compiled libs, etc. |
| `docker-build` | Build Linux libs in Docker | Uses CentOS 7 image for glibc compatibility |

### Automatic CGO Environment Setup

`build-go` and `test` subcommands auto-set environment before running `go build`/`go test`:

```go
func setupCGOEnv() error {
    libDir := filepath.Join("lib", runtime.GOOS + "-" + arch)
    includeDir := "include"
    os.Setenv("CGO_CFLAGS", "-I" + absPath(includeDir))
    os.Setenv("CGO_CXXFLAGS", "-std=c++17 -I" + absPath(includeDir))
    os.Setenv("CGO_LDFLAGS", "-L" + absPath(libDir) + " -lzxingwrapper -lZXing -lstdc++ -lm")
    os.Setenv("CGO_ENABLED", "1")
    return nil
}
```

### Cross-platform Implementation

`cmd/build/` is pure Go, calls cmake/go/docker via `os/exec`. No shell scripts needed; Windows/Linux/macOS unified behavior.

## Directory Structure

```
zxing/
├── cmd/
│   ├── zxing-cli/          # CLI tool (retained)
│   ├── server/             # HTTP service (retained)
│   ├── wasm-example/       # WASM example (retained)
│   └── build/              # NEW: cross-platform build tool
├── pkg/
│   ├── zxing/              # Unified interface layer
│   └── wasm/               # WASM runtime (wazero implementation)
├── lib/                    # Pre-compiled static libraries
│   ├── linux-x64/
│   │   ├── libZXing.a
│   │   └── libzxingwrapper.a
│   └── windows-x64/
│       ├── ZXing.lib
│       └── zxingwrapper.lib
├── include/                # C header files
│   ├── zxing.h
│   ├── zxing_internal.h
│   └── ZXing/              # zxing-cpp public headers
├── wasm/                   # Pre-compiled WASM file
│   └── zxingwrapper.wasm
├── zxing-cpp/              # git submodule (v2.3.0)
├── src/                    # C++ wrapper source
│   ├── zxing.cpp
│   └── zxing_internal.h
├── CMakeLists.txt          # CMake build config
├── CMakeLists-wasm.txt     # WASM CMake config
├── docker/                 # Docker build environment
│   └── Dockerfile.linux-build
├── go.mod
└── README.md
```

## Files to Delete

| File | Replaced By |
|------|-------------|
| `build.sh` | `go run ./cmd/build build-go` |
| `build.bat` | `go run ./cmd/build build-go` |
| `build.ps1` | `go run ./cmd/build build-go` |
| `build_all.ps1` | `go run ./cmd/build build-all` |
| `build_wasm.ps1` | `go run ./cmd/build build-wasm` |
| `build_wasm.sh` | `go run ./cmd/build build-wasm` |
| `build_wasm_demo.ps1` | Deleted (wasm-example retained, no separate build script) |
| `scripts/build_static_linux.sh` | `go run ./cmd/build build-lib` or `docker-build` |
| `scripts/build_static_windows.ps1` | `go run ./cmd/build build-lib` |
| `scripts/build_wasm_save.sh` | `go run ./cmd/build build-wasm` |
| `scripts/build_wasm_save.ps1` | `go run ./cmd/build build-wasm` |
| `test_build.ps1` | `go run ./cmd/build test` |
| `test_integration.ps1` | `go run ./cmd/build test` |
| `test_cmake.sh` | `go run ./cmd/build test` |
| `Makefile` | `go run ./cmd/build` |
| `lib64/` | Consolidated into `lib/` |

## Testing Strategy

### Test Types

| Test Type | Scope | Command |
|-----------|-------|---------|
| Unit test | `pkg/zxing/` interface, types, config | `go run ./cmd/build test` |
| CGO integration test | Real image decode | `CGO_ENABLED=1 go test ./pkg/zxing/ -v -run TestDecode` |
| WASM integration test | wazero WASM decode | `CGO_ENABLED=0 go test ./pkg/zxing/ -v -run TestDecode` |
| Build smoke test | Verify `go build` in both modes | CI |

### Test File Adjustments

- `TestBackendSelection` — verify compile-time backend availability (CGO mode returns `BackendCGO`, WASM mode returns `BackendWASM`)
- `TestDecodeBytes` — use real QR Code test image (from `test/` directory), verify expected text
- New `TestEncodeDecode` — encode then decode round-trip (WASM backend only, CGO doesn't support encoding)

### CI Verification Matrix

| Environment | CGO_ENABLED | Expected Backend | Verification |
|-------------|-------------|------------------|-------------|
| Linux (glibc 2.17+) | 1 | CGO | go build + go test |
| Linux (glibc 2.17+) | 0 | WASM | go build + go test |
| Windows x64 | 1 | CGO | go build + go test |
| Windows x64 | 0 | WASM | go build + go test |
| GOOS=js GOARCH=wasm | - | WASM(js) | go build |

## README Structure

1. Intro + features
2. Quick start (pre-compiled libs — clone and go build)
3. Backend selection table
4. Usage examples (CGO and unified interface)
5. CLI tool usage
6. Build tool commands (`cmd/build/` subcommands)
7. Building from source (Docker for Linux, local for Windows, EMSDK for WASM)
8. Project structure
9. API documentation
10. Testing
11. License

## Implementation Order

1. Restructure `pkg/zxing/` — split CGO bindings by platform, add build tags
2. Implement wazero WASM runtime in `pkg/wasm/`
3. Simplify `universal_impl.go` / `factory.go` — compile-time backend selection
4. Create `cmd/build/` — cross-platform build tool
5. Create `docker/Dockerfile.linux-build` — CentOS 7 build environment
6. Build and commit pre-compiled libraries to `lib/`
7. Delete old build scripts
8. Update README.md
9. Update tests
10. Verify: `CGO_ENABLED=1 go build` and `CGO_ENABLED=0 go build` both work
