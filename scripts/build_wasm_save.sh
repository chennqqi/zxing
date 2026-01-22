#!/bin/bash
# WASM构建并保存脚本 (Linux)
# 编译ZXingCPP为WASM，并保存到wasm目录

set -e

BUILD_TYPE=${1:-Release}

echo "========================================"
echo "Building WASM Module for Linux"
echo "========================================"
echo "Build Type: $BUILD_TYPE"
echo ""

# 检查Emscripten SDK
if [ -z "$EMSDK" ]; then
    echo "Error: EMSDK environment variable not set"
    echo "Please install Emscripten SDK first:"
    echo "  https://emscripten.org/docs/getting_started/downloads.html"
    exit 1
fi

echo "Using Emscripten SDK at: $EMSDK"

# 检查emcc命令
if ! command -v emcc &> /dev/null; then
    echo "Error: emcc command not found"
    echo "Please activate Emscripten SDK first:"
    echo "  source $EMSDK/emsdk_env.sh"
    exit 1
fi

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 检查并下载ZXingCPP源码
if [ ! -d "zxing-cpp" ]; then
    echo "Downloading ZXingCPP source code..."
    if [ -n "$GIT_PROXY" ]; then
        echo "Using proxy $GIT_PROXY for git clone"
        https_proxy="$GIT_PROXY" git clone https://github.com/zxing-cpp/zxing-cpp.git --recursive --single-branch --depth 1
    else
        git clone https://github.com/zxing-cpp/zxing-cpp.git --recursive --single-branch --depth 1
    fi
    if [ $? -ne 0 ]; then
        echo "Error: Failed to clone ZXingCPP repository"
        exit 1
    fi
fi

# 创建构建目录
BUILD_DIR="build-wasm"
if [ -d "$BUILD_DIR" ]; then
    echo "Removing existing build directory..."
    rm -rf "$BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# 配置CMake - 使用WASM工具链
echo "Configuring CMake for WASM build..."
TOOLCHAIN_FILE="$EMSDK/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake"

if [ ! -f "$TOOLCHAIN_FILE" ]; then
    echo "Error: Emscripten toolchain file not found: $TOOLCHAIN_FILE"
    exit 1
fi

# 使用CMakeLists-wasm.txt作为配置
cp "../CMakeLists-wasm.txt" "../CMakeLists.txt.backup" 2>/dev/null || true
cp "../CMakeLists-wasm.txt" "../CMakeLists.txt"

cmake .. \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DCMAKE_TOOLCHAIN_FILE="$TOOLCHAIN_FILE"

if [ $? -ne 0 ]; then
    echo "Error: CMake configuration failed"
    # 恢复原始CMakeLists.txt
    if [ -f "../CMakeLists.txt.backup" ]; then
        mv "../CMakeLists.txt.backup" "../CMakeLists.txt"
    fi
    exit 1
fi

# 恢复原始CMakeLists.txt
if [ -f "../CMakeLists.txt.backup" ]; then
    mv "../CMakeLists.txt.backup" "../CMakeLists.txt"
fi

# 构建
echo "Building WASM module..."
cmake --build . --config "$BUILD_TYPE" -j$(nproc)

if [ $? -ne 0 ]; then
    echo "Error: Build failed"
    exit 1
fi

# 创建wasm目录
WASM_DIR="$PROJECT_ROOT/wasm"
mkdir -p "$WASM_DIR"

# 复制WASM文件
echo "Copying WASM files..."

# 查找WASM文件
WASM_FILE=$(find ./bin -name "*.wasm" -type f 2>/dev/null | head -n 1)
if [ -n "$WASM_FILE" ] && [ -f "$WASM_FILE" ]; then
    cp "$WASM_FILE" "$WASM_DIR/zxing.wasm"
    echo "  Copied: zxing.wasm"
else
    echo "  Warning: WASM file not found"
fi

# 查找JS文件
JS_FILE=$(find ./bin -name "*.js" -type f 2>/dev/null | head -n 1)
if [ -n "$JS_FILE" ] && [ -f "$JS_FILE" ]; then
    cp "$JS_FILE" "$WASM_DIR/zxing.js"
    echo "  Copied: zxing.js"
else
    echo "  Warning: JS file not found"
fi

# 验证文件
echo ""
echo "Verifying WASM files..."
if [ -f "$WASM_DIR/zxing.wasm" ]; then
    SIZE=$(du -h "$WASM_DIR/zxing.wasm" | cut -f1)
    echo "  zxing.wasm: $SIZE"
else
    echo "  zxing.wasm: NOT FOUND"
fi

if [ -f "$WASM_DIR/zxing.js" ]; then
    SIZE=$(du -h "$WASM_DIR/zxing.js" | cut -f1)
    echo "  zxing.js: $SIZE"
else
    echo "  zxing.js: NOT FOUND"
fi

echo ""
echo "========================================"
echo "WASM build completed successfully!"
echo "Files saved to: $WASM_DIR"
echo "========================================"
