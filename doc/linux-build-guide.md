# Linux Environment Build Guide

## Prerequisites

### System Requirements
- Linux (tested on Rocky Linux/CentOS/Ubuntu)
- CMake 3.10+
- GCC/G++ compiler with C++20 support
- Go 1.18+
- Git

### Install Dependencies

#### Rocky Linux/CentOS
```bash
# Install development tools
sudo yum groupinstall "Development Tools"
sudo yum install cmake git golang

# Check versions
cmake --version
gcc --version
go version
```

#### Ubuntu/Debian
```bash
# Install development tools
sudo apt update
sudo apt install build-essential cmake git golang-go

# Check versions
cmake --version
gcc --version
go version
```

## Build CGO Version

### Step 1: Check zxing-cpp submodule
```bash
cd /data/dev/github.com/chennqqi/zxing

# Check if zxing-cpp submodule exists
ls -la zxing-cpp/

# If empty, initialize submodule
git submodule update --init --recursive
```

### Step 2: Build CGO version
```bash
# Make build script executable
chmod +x build.sh

# Build (this will automatically build zxing-cpp if needed)
./build.sh
```

### Step 3: Verify CGO build
```bash
# Check static library
ls -lh lib/

# Test CLI
./bin/zxing-cli --help

# Test with a QR code image (if you have one)
# ./bin/zxing-cli decode testdata/qrcode.png
```

## Build WASM Version

### Step 1: Install Emscripten SDK
```bash
# Clone emsdk (if not already installed)
cd ~
git clone https://github.com/emscripten-core/emsdk.git
cd emsdk

# Install and activate latest version
./emsdk install latest
./emsdk activate latest

# Setup environment
source ./emsdk_env.sh

# Verify installation
emcc --version
```

### Step 2: Build WASM version
```bash
cd /data/dev/github.com/chennqqi/zxing

# Make build script executable
chmod +x build_wasm.sh

# Activate emsdk environment (if not already done)
source ~/emsdk/emsdk_env.sh

# Build WASM
./build_wasm.sh
```

### Step 3: Verify WASM build
```bash
# Check WASM files
ls -lh wasm/zxingwrapper.*

# Check file sizes (should be ~60KB for .js and ~680KB for .wasm)
du -h wasm/zxingwrapper.js
du -h wasm/zxingwrapper.wasm
```

## Troubleshooting

### Error: CMake version too old
```bash
# Install newer CMake from official website
wget https://github.com/Kitware/CMake/releases/download/v3.27.0/cmake-3.27.0-linux-x86_64.sh
sudo sh cmake-3.27.0-linux-x86_64.sh --prefix=/usr/local --skip-license
```

### Error: GCC version too old (need C++20)
```bash
# Rocky Linux/CentOS - install devtoolset
sudo yum install centos-release-scl
sudo yum install devtoolset-11-gcc devtoolset-11-gcc-c++
scl enable devtoolset-11 bash

# Ubuntu - install newer GCC
sudo apt install gcc-11 g++-11
sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-11 100
sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-11 100
```

### Error: Cannot find zxing-cpp
```bash
# Force rebuild from source
ZXING_INSTALL=1 ./build.sh
```

### Error: Emscripten not found
```bash
# Make sure emsdk environment is activated
source ~/emsdk/emsdk_env.sh

# Verify
echo $EMSDK
which emcc
```

## Testing Both Versions

### Test CGO version
```bash
# Build and test
go test ./... -v

# Run benchmark
go test -bench=. ./...
```

### Test WASM version
```bash
# WASM tests (if implemented)
# go test -tags wasm ./... -v
```

## Cross-Verification with Windows

After building on Linux, you should:

1. Commit the Linux build artifacts (if different from Windows):
   ```bash
   git status
   git add lib/ wasm/
   git commit -m "Add Linux build artifacts"
   ```

2. Return to Windows and verify no conflicts
3. Run Windows builds to ensure no regression

## Build Output

After successful build, you should have:

### CGO Build Output
- `lib/libzxingwrapper.a` - Static library
- `bin/zxing-cli` - CLI executable

### WASM Build Output
- `wasm/zxingwrapper.js` - JavaScript loader (60KB)
- `wasm/zxingwrapper.wasm` - WASM binary (~680KB)

## Notes

1. **CMakeLists.txt Management**: The build scripts automatically manage switching between CGO and WASM configurations
2. **Parallel Builds**: Both CGO and WASM can be built on the same system
3. **Clean Build**: Use `rm -rf build/ build-wasm/` to start fresh if needed
4. **Submodule**: The zxing-cpp submodule is shared between Windows and Linux builds
