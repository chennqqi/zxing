#!/bin/bash

# WASM构建脚本 for Linux
# 需要先安装Emscripten SDK

set -e

BUILD_TYPE=${1:-Release}
BUILD_DIR=${2:-build-wasm}

echo "Building ZXing WASM module..."

# 检查Emscripten是否安装
if [ -z "$EMSDK" ]; then
    echo "Error: EMSDK environment variable not set. Please install Emscripten SDK first."
    echo "Installation guide: https://emscripten.org/docs/getting_started/downloads.html"
    exit 1
fi

echo "Using Emscripten SDK at: $EMSDK"

# 检查Emscripten是否可用
if ! command -v emcc &> /dev/null; then
    echo "Error: emcc command not found. Please activate Emscripten SDK first:"
    echo "source $EMSDK/emsdk_env.sh"
    exit 1
fi

# 创建构建目录
if [ -d "$BUILD_DIR" ]; then
    echo "Removing existing build directory..."
    rm -rf "$BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"

# 进入构建目录
cd "$BUILD_DIR"

# 配置CMake
echo "Configuring CMake..."
cmake \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DCMAKE_TOOLCHAIN_FILE="$EMSDK/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake" \
    ..

# 构建项目
echo "Building project..."
cmake --build . --config "$BUILD_TYPE"

# 检查输出文件
WASM_FILE="bin/zxingwrapper.wasm"
JS_FILE="bin/zxingwrapper.js"

if [ -f "$WASM_FILE" ]; then
    WASM_SIZE=$(du -h "$WASM_FILE" | cut -f1)
    echo "WASM file generated: $WASM_FILE ($WASM_SIZE)"
else
    echo "Warning: WASM file not found!"
fi

if [ -f "$JS_FILE" ]; then
    JS_SIZE=$(du -h "$JS_FILE" | cut -f1)
    echo "JS file generated: $JS_FILE ($JS_SIZE)"
else
    echo "Warning: JS file not found!"
fi

# 复制文件到项目目录
echo "Copying files to project directory..."
WASM_DIR="../wasm"
mkdir -p "$WASM_DIR"

if [ -f "$WASM_FILE" ]; then
    cp "$WASM_FILE" "$WASM_DIR/"
    echo "Copied WASM file to: $WASM_DIR"
fi

if [ -f "$JS_FILE" ]; then
    cp "$JS_FILE" "$WASM_DIR/"
    echo "Copied JS file to: $WASM_DIR"
fi

echo "WASM build completed successfully!"

# 返回原目录
cd ..
