# AGENTS.md

Guidelines for AI agents and contributors working on this repository.

## Linux Static Library Build — glibc Compatibility

**CRITICAL**: Linux static libraries (`lib/linux-x64/`) MUST be built on a glibc-based system with glibc <= 2.17 to ensure broad binary compatibility.

### Rule

Never use Alpine Linux (musl libc) or any musl-based distribution to build Linux static libraries. musl-compiled binaries are **not compatible** with glibc-based systems (Ubuntu, CentOS, RHEL, Debian, etc.).

### Approved Build Environment

- **Base image**: CentOS 7 (glibc 2.17)
- **Toolchain**: devtoolset-10 (GCC 10.2.1, C++20 partial support)
- **CMake**: cmake3 >= 3.17 (from EPEL)
- **Dockerfile**: `docker/Dockerfile.linux-build`
- **Patch script**: `docker/patch_using_enum.sh` (required for GCC 10 compatibility)

### Why CentOS 7 / glibc 2.17?

glibc is backward compatible: a binary compiled against glibc 2.17 runs on any system with glibc >= 2.17. This covers:

- CentOS / RHEL 7+
- Ubuntu 16.04+
- Debian 8+
- Amazon Linux 2+
- All major Linux distributions released after 2014

### GCC 10 Patches

GCC 10 does not support two C++20 features used by zxing-cpp v3.0.2:

1. **`using enum`** (P0648R2, GCC 11+): Patched by `docker/patch_using_enum.sh`, which replaces `using enum BarcodeFormat;` with `static constexpr auto` declarations.
2. **Coroutines**: Enabled via `-DCMAKE_CXX_FLAGS=-fcoroutines` CMake flag.

### Build Command

**Static libraries only:**

```bash
docker build -t zxing-linux-build -f docker/Dockerfile.linux-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-linux-build \
  sh -c "/tmp/patch_using_enum.sh /workspace/zxing-cpp && \
  cd /tmp && rm -rf build && mkdir -p build && cd build && \
  cmake3 -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF \
  -DCMAKE_CXX_STANDARD=20 -DCMAKE_CXX_FLAGS=-fcoroutines /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/"
```

**Full CGO binary (static libs + Go binary, glibc 2.17 compatible):**

The Docker image includes Go 1.24. The entire `go build` must run inside the container to ensure the linker uses glibc 2.17 version symbols.

```bash
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

**Why `go build` must run inside the container:**

Static libraries (.a) have no glibc version markers. But when `go build` links the final binary, the linker resolves undefined symbols against the **host system's** glibc, inheriting its version tags. Building on a host with glibc 2.39 produces a binary requiring GLIBC_2.38, even if the static libs were built on CentOS 7. Building entirely inside CentOS 7 ensures the highest GLIBC requirement is 2.14.

### Windows Cross-Compile

Windows static libraries (`lib/windows-x64/`) are built with MinGW-w64 cross-compiler. Alpine is acceptable for Windows cross-compilation because MinGW produces PE binaries (Windows format), not ELF. The host libc (musl) does not affect the output.

- **Dockerfile**: `docker/Dockerfile.win-build` (Alpine 3.18 + MinGW-w64)

## WASM Module

The WASM module (`wasm/zxingwrapper.wasm`) is platform-independent. A single `.wasm` file runs on all platforms (Linux, Windows, macOS, browsers) via wazero or `GOOS=js GOARCH=wasm`.

## Build Tags

CGO backend is available on `linux` and `windows` only. WASM backend is used on all other platforms or when `CGO_ENABLED=0`.

| Condition | Backend |
|-----------|---------|
| `CGO_ENABLED=1` + Linux/Windows | CGO |
| `CGO_ENABLED=0` or macOS | WASM (wazero) |
| `GOOS=js GOARCH=wasm` | WASM (js) |

## Code Style

- Go code: use `len(strVal) == 0` instead of `strVal == ""` for string comparison
- Comments, documentation, and log messages: English
- JSON keys: lowercase with underscores (snake_case)
- New functions and classes: must have doc comments
- Use early return pattern
- File encoding: UTF-8
