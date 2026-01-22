#!/bin/bash
# 验证lib目录中的静态库文件

set -e

echo "========================================"
echo "Verifying Library Files"
echo "========================================"
echo ""

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

ALL_FOUND=true

# 检查Windows静态库
echo "Checking Windows static libraries..."
WINDOWS_LIB_DIR="$PROJECT_ROOT/lib/windows/x64"

if [ -d "$WINDOWS_LIB_DIR" ]; then
    if [ -f "$WINDOWS_LIB_DIR/ZXing.lib" ]; then
        SIZE=$(du -h "$WINDOWS_LIB_DIR/ZXing.lib" | cut -f1)
        echo "  [OK] ZXing.lib: $SIZE"
    else
        echo "  [MISSING] ZXing.lib"
        ALL_FOUND=false
    fi
    
    if [ -f "$WINDOWS_LIB_DIR/zxingwrapper.lib" ]; then
        SIZE=$(du -h "$WINDOWS_LIB_DIR/zxingwrapper.lib" | cut -f1)
        echo "  [OK] zxingwrapper.lib: $SIZE"
    else
        echo "  [MISSING] zxingwrapper.lib"
        ALL_FOUND=false
    fi
else
    echo "  [MISSING] Directory: $WINDOWS_LIB_DIR"
    ALL_FOUND=false
fi

echo ""

# 检查Linux静态库
echo "Checking Linux static libraries..."
LINUX_LIB_DIR="$PROJECT_ROOT/lib/linux/x64"

if [ -d "$LINUX_LIB_DIR" ]; then
    if [ -f "$LINUX_LIB_DIR/libZXing.a" ]; then
        SIZE=$(du -h "$LINUX_LIB_DIR/libZXing.a" | cut -f1)
        echo "  [OK] libZXing.a: $SIZE"
    else
        echo "  [MISSING] libZXing.a"
        ALL_FOUND=false
    fi
    
    if [ -f "$LINUX_LIB_DIR/libzxingwrapper.a" ]; then
        SIZE=$(du -h "$LINUX_LIB_DIR/libzxingwrapper.a" | cut -f1)
        echo "  [OK] libzxingwrapper.a: $SIZE"
    else
        echo "  [MISSING] libzxingwrapper.a"
        ALL_FOUND=false
    fi
else
    echo "  [MISSING] Directory: $LINUX_LIB_DIR"
    ALL_FOUND=false
fi

echo ""

# 检查WASM文件
echo "Checking WASM files..."
WASM_DIR="$PROJECT_ROOT/wasm"

if [ -d "$WASM_DIR" ]; then
    if [ -f "$WASM_DIR/zxing.wasm" ]; then
        SIZE=$(du -h "$WASM_DIR/zxing.wasm" | cut -f1)
        echo "  [OK] zxing.wasm: $SIZE"
    else
        echo "  [MISSING] zxing.wasm"
        ALL_FOUND=false
    fi
    
    if [ -f "$WASM_DIR/zxing.js" ]; then
        SIZE=$(du -h "$WASM_DIR/zxing.js" | cut -f1)
        echo "  [OK] zxing.js: $SIZE"
    else
        echo "  [MISSING] zxing.js"
        ALL_FOUND=false
    fi
else
    echo "  [MISSING] Directory: $WASM_DIR"
    ALL_FOUND=false
fi

echo ""
echo "========================================"
if [ "$ALL_FOUND" = true ]; then
    echo "All library files are present!"
    exit 0
else
    echo "Some library files are missing!"
    echo "Please run the build scripts to generate them."
    exit 1
fi
