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

    # Initialize zxing-cpp submodule
    echo "Initializing zxing-cpp submodule..."
    if [ -n "$GIT_PROXY" ]; then
        https_proxy="$GIT_PROXY" git submodule update --init --recursive
    else
        git submodule update --init --recursive
    fi

# 创建构建目录
mkdir -p build
cd build

# 配置 CMake
echo "Configuring CMake..."
cmake .. -DCMAKE_INSTALL_PREFIX=.. -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_LIBDIR=lib -DBUILD_SHARED_LIBS=OFF
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

# Copy ZXing static library manually since it's not installed by default
if [ -f "zxing-cpp/core/libZXing.a" ]; then
    cp zxing-cpp/core/libZXing.a ../lib/
    echo "Copied libZXing.a to lib/"
fi

cd ..

# 设置 CGO 环境变量
export CGO_CFLAGS="-I$(pwd)/include -I$(pwd)/zxing-cpp/core/src"
export CGO_CXXFLAGS="-std=c++17 -I$(pwd)/include -I$(pwd)/zxing-cpp/core/src"
export CGO_LDFLAGS="-L$(pwd)/lib -lzxingwrapper -lZXing -lstdc++ -lm"

# 检查 include 目录内容
echo "Checking include directory:"
ls -l include/zxing.h
ls -l include/ZXing/

echo "Building Go CLI..."
go build -o bin/zxing-cli ./cmd/zxing-cli/
if [ $? -ne 0 ]; then
    echo "Error: Go build failed"
    exit 1
fi

echo "Build completed successfully!"
echo "Static library installed to: lib/"
echo "CLI binary installed to: bin/zxing-cli"