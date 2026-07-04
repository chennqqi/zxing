#!/bin/bash
# Linux静态库编译脚本
# 编译ZXingCPP和zxingwrapper为静态库，并保存到lib目录

set -e

BUILD_TYPE=${1:-Release}
ARCH=${2:-x64}

echo "========================================"
echo "Building Static Libraries for Linux"
echo "========================================"
echo "Build Type: $BUILD_TYPE"
echo "Architecture: $ARCH"
echo ""

# 检查依赖
echo "Checking dependencies..."
if ! command -v cmake &> /dev/null; then
    echo "Error: CMake is not installed"
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo "Error: Git is not installed"
    exit 1
fi

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Initialize zxing-cpp submodule
echo "Initializing zxing-cpp submodule..."
if [ -n "$GIT_PROXY" ]; then
    https_proxy="$GIT_PROXY" git submodule update --init --recursive
else
    git submodule update --init --recursive
fi
if [ $? -ne 0 ]; then
    echo "Error: Failed to initialize zxing-cpp submodule"
    exit 1
fi

# 创建构建目录
BUILD_DIR="build-static-linux"
if [ -d "$BUILD_DIR" ]; then
    echo "Removing existing build directory..."
    rm -rf "$BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# 配置CMake - 编译静态库
echo "Configuring CMake for static library build..."
cmake .. \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DBUILD_STATIC_LIB=ON \
    -DBUILD_SHARED_LIB=OFF \
    -DCMAKE_INSTALL_PREFIX="$PROJECT_ROOT"

if [ $? -ne 0 ]; then
    echo "Error: CMake configuration failed"
    exit 1
fi

# 构建
echo "Building static libraries..."
cmake --build . --config "$BUILD_TYPE" -j$(nproc)

if [ $? -ne 0 ]; then
    echo "Error: Build failed"
    exit 1
fi

# 创建lib目录结构
LIB_DIR="$PROJECT_ROOT/lib/linux/$ARCH"
mkdir -p "$LIB_DIR"

# 复制静态库文件
echo "Copying static library files..."

# 查找ZXing静态库
ZXING_LIB=$(find ./lib -name "libZXing.a" -type f 2>/dev/null | head -n 1)
if [ -n "$ZXING_LIB" ] && [ -f "$ZXING_LIB" ]; then
    cp "$ZXING_LIB" "$LIB_DIR/libZXing.a"
    echo "  Copied: libZXing.a"
else
    echo "  Warning: libZXing.a not found"
fi

# 查找zxingwrapper静态库
WRAPPER_LIB=$(find ./lib -name "libzxingwrapper.a" -type f 2>/dev/null | head -n 1)
if [ -n "$WRAPPER_LIB" ] && [ -f "$WRAPPER_LIB" ]; then
    cp "$WRAPPER_LIB" "$LIB_DIR/libzxingwrapper.a"
    echo "  Copied: libzxingwrapper.a"
else
    echo "  Warning: libzxingwrapper.a not found"
fi

# 验证文件
echo ""
echo "Verifying library files..."
if [ -f "$LIB_DIR/libZXing.a" ]; then
    SIZE=$(du -h "$LIB_DIR/libZXing.a" | cut -f1)
    echo "  libZXing.a: $SIZE"
else
    echo "  libZXing.a: NOT FOUND"
fi

if [ -f "$LIB_DIR/libzxingwrapper.a" ]; then
    SIZE=$(du -h "$LIB_DIR/libzxingwrapper.a" | cut -f1)
    echo "  libzxingwrapper.a: $SIZE"
else
    echo "  libzxingwrapper.a: NOT FOUND"
fi

echo ""
echo "========================================"
echo "Build completed successfully!"
echo "Libraries saved to: $LIB_DIR"
echo "========================================"
