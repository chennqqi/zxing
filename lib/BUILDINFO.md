# Pre-compiled Libraries

## Linux x64 (`lib/linux-x64/`)

- `libZXing.a` — ZXing-CPP static library
- `libzxingwrapper.a` — C wrapper static library

### Build Environment

- **OS**: Linux x86_64
- **Compiler**: GCC (C++17)
- **ZXing-CPP version**: v2.3.0 (git submodule)
- **Build type**: Release

### Build Instructions

```bash
# Using the build tool (requires Docker)
go run ./cmd/build docker-build

# Or build locally
go run ./cmd/build build-lib
```

### Compatibility

For maximum glibc compatibility (glibc 2.17, CentOS 7), use the Docker build:

```bash
go run ./cmd/build docker-build
```

This produces libraries compatible with:
- CentOS 7+ (glibc 2.17+)
- Ubuntu 18.04+ (glibc 2.27+)
- Debian 10+ (glibc 2.28+)
- Most modern Linux distributions

## Windows x64 (`lib/windows-x64/`)

To build Windows libraries, use MinGW-w64 cross-compilation or build natively on Windows:

```bash
# On Windows with MinGW
go run ./cmd/build build-lib
```

## WASM (`wasm/zxingwrapper.wasm`)

The WASM module is built using Emscripten with `STANDALONE_WASM=1` for wazero compatibility:

```bash
go run ./cmd/build build-wasm
```
