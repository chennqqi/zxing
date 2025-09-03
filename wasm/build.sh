#!/bin/bash

# ZXing WASM 构建脚本
# 使用 Emscripten 编译简化版 zxing 为 WebAssembly

set -e

echo "开始构建 ZXing WASM 模块..."

# 检查 Emscripten 环境
if ! command -v emcc &> /dev/null; then
    echo "错误: 未找到 Emscripten 编译器"
    echo "请先安装 Emscripten SDK:"
    echo "  git clone https://github.com/emscripten-core/emsdk.git"
    echo "  cd emsdk"
    echo "  ./emsdk install latest"
    echo "  ./emsdk activate latest"
    echo "  source ./emsdk_env.sh"
    exit 1
fi

# 创建构建目录
BUILD_DIR="build"
mkdir -p $BUILD_DIR

echo "编译 C++ 源码为 WASM..."

# 编译简化版 zxing 为 WASM
emcc \
    -O3 \
    -s WASM=1 \
    -s EXPORTED_RUNTIME_METHODS='["ccall", "cwrap", "getValue", "setValue"]' \
    -s EXPORTED_FUNCTIONS='["_decode_image_data", "_encode_text_to_qr", "_free_decode_result", "_free_encode_result", "_malloc", "_free"]' \
    -s ALLOW_MEMORY_GROWTH=1 \
    -s MODULARIZE=1 \
    -s EXPORT_NAME='ZXingWASM' \
    -s ENVIRONMENT='web,worker' \
    --bind \
    -std=c++17 \
    zxing_simple.cpp \
    -o $BUILD_DIR/zxing.js

if [ $? -eq 0 ]; then
    echo "WASM 编译成功！"
    
    # 复制生成的文件
    cp $BUILD_DIR/zxing.js ./
    cp $BUILD_DIR/zxing.wasm ./
    
    echo "生成的文件:"
    echo "  - zxing.js (JavaScript 加载器)"
    echo "  - zxing.wasm (WebAssembly 模块)"
    
    # 获取文件大小
    WASM_SIZE=$(stat -f%z zxing.wasm 2>/dev/null || stat -c%s zxing.wasm 2>/dev/null || echo "unknown")
    echo "  - WASM 文件大小: $WASM_SIZE bytes"
    
else
    echo "WASM 编译失败！"
    exit 1
fi

echo "构建完成！"