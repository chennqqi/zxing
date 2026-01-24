#!/bin/bash
# Linux Quick Start - Execute this in your SSH terminal
# Location: /data/dev/github.com/chennqqi/zxing

set -e  # Exit on error

echo "================================"
echo "ZXing Linux Build Quick Start"
echo "================================"
echo ""

# Step 1: Pull latest changes
echo "[1/8] Pulling latest changes from git..."
git pull

# Step 2: Check environment
echo ""
echo "[2/8] Checking build environment..."
echo -n "CMake version: "
cmake --version | head -n1
echo -n "GCC version: "
gcc --version | head -n1
echo -n "Go version: "
go version

# Step 3: Check zxing-cpp submodule
echo ""
echo "[3/8] Checking zxing-cpp submodule..."
if [ ! -f "zxing-cpp/CMakeLists.txt" ]; then
    echo "Initializing zxing-cpp submodule..."
    git submodule update --init --recursive
else
    echo "zxing-cpp submodule already initialized"
fi

# Step 4: Make scripts executable
echo ""
echo "[4/8] Making build scripts executable..."
chmod +x build.sh build_wasm.sh

# Step 5: Build CGO version
echo ""
echo "[5/8] Building CGO version..."
echo "This may take a few minutes..."
./build.sh

# Step 6: Verify CGO build
echo ""
echo "[6/8] Verifying CGO build..."
if [ -f "lib/libzxingwrapper.a" ]; then
    echo "✓ Static library built: $(ls -lh lib/libzxingwrapper.a | awk '{print $5}')"
else
    echo "✗ Static library not found!"
    exit 1
fi

if [ -f "bin/zxing-cli" ]; then
    echo "✓ CLI binary built: $(ls -lh bin/zxing-cli | awk '{print $5}')"
    echo "Testing CLI..."
    ./bin/zxing-cli --help > /dev/null 2>&1 && echo "✓ CLI works!" || echo "✗ CLI test failed"
else
    echo "✗ CLI binary not found!"
    exit 1
fi

# Step 7: Build WASM version (optional - requires emsdk)
echo ""
echo "[7/8] Building WASM version..."
if [ -n "$EMSDK" ]; then
    echo "Emscripten SDK found at: $EMSDK"
    ./build_wasm.sh
    
    # Verify WASM build
    echo ""
    echo "Verifying WASM build..."
    if [ -f "wasm/zxingwrapper.wasm" ]; then
        echo "✓ WASM module built: $(ls -lh wasm/zxingwrapper.wasm | awk '{print $5}')"
    else
        echo "✗ WASM module not found!"
    fi
    
    if [ -f "wasm/zxingwrapper.js" ]; then
        echo "✓ WASM loader built: $(ls -lh wasm/zxingwrapper.js | awk '{print $5}')"
    else
        echo "✗ WASM loader not found!"
    fi
else
    echo "⚠ Emscripten SDK not found. Skipping WASM build."
    echo "To build WASM, install emsdk and run:"
    echo "  source ~/emsdk/emsdk_env.sh"
    echo "  ./build_wasm.sh"
fi

# Step 8: Summary
echo ""
echo "[8/8] Build Summary"
echo "================================"
echo "Build completed successfully!"
echo ""
echo "Generated files:"
echo "  CGO:"
ls -lh lib/ bin/zxing-cli 2>/dev/null || echo "    (none)"
echo ""
echo "  WASM:"
ls -lh wasm/zxingwrapper.* 2>/dev/null || echo "    (not built)"
echo ""
echo "Next steps:"
echo "  1. Test CLI: ./bin/zxing-cli --help"
echo "  2. Run tests: go test ./..."
echo "  3. Compare with Windows build artifacts"
echo ""
echo "================================"
