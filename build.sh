#!/bin/bash

# 检查是否安装了 CMake
if ! command -v cmake &> /dev/null; then
    echo "Error: CMake is not installed"
    exit 1
fi

# 检查是否安装了 Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# 创建构建目录
mkdir -p build
cd build

# 配置 CMake
echo "Configuring CMake..."
cmake .. \
    -DCMAKE_INSTALL_PREFIX=.. \
    -DCMAKE_BUILD_TYPE=Release

if [ $? -ne 0 ]; then
    echo "Error: CMake configuration failed"
    exit 1
fi

# 构建
echo "Building..."
cmake --build . --config Release

if [ $? -ne 0 ]; then
    echo "Error: Build failed"
    exit 1
fi

# 安装
echo "Installing..."
cmake --install . --config Release

if [ $? -ne 0 ]; then
    echo "Error: Installation failed"
    exit 1
fi

# 返回上级目录
cd ..

# 构建 Go 库
echo "Building Go library..."

# 根据操作系统选择输出文件名
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    go build -o bin/libzxing.dylib -buildmode=c-shared
else
    # Linux
    go build -o bin/libzxing.so -buildmode=c-shared
fi

if [ $? -ne 0 ]; then
    echo "Error: Go build failed"
    exit 1
fi

echo "Build completed successfully!" 