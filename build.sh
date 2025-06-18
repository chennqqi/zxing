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

# 检查是否需要强制源码安装 ZXing
force_src=${ZXING_INSTALL:-0}

# 检查系统是否已安装 ZXing（通过 pkg-config）
pkg-config --exists zxing
has_zxing=$?

# 检查系统ZXing版本
zxing_ver=""
if [ $has_zxing -eq 0 ]; then
    zxing_ver=$(pkg-config --modversion zxing)
    # 版本号小于2.0.0则强制源码
    if [ "$(echo -e "$zxing_ver\n2.0.0" | sort -V | head -n1)" = "$zxing_ver" ] && [ "$zxing_ver" != "2.0.0" ]; then
        echo "[build.sh] ZXing version $zxing_ver is too old, will build from source."
        has_zxing=1
    fi
fi

if [ "$force_src" = "1" ] || [ $has_zxing -ne 0 ]; then
    echo "[build.sh] ZXing not found or forced to build from source, building from source..."
    if [ ! -d "zxing-cpp" ]; then
        if [ -n "$GIT_PROXY" ]; then
            echo "[build.sh] Using proxy $GIT_PROXY for git clone"
            https_proxy="$GIT_PROXY" git clone https://github.com/zxing-cpp/zxing-cpp.git --recursive --single-branch --depth 1
        else
            git clone https://github.com/zxing-cpp/zxing-cpp.git --recursive --single-branch --depth 1
        fi
    fi
    cd zxing-cpp
    mkdir -p build && cd build
    cmake -S .. -B . -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIB=ON
    if [ $? -ne 0 ]; then
        echo "Error: ZXing CMake configuration failed"
        exit 1
    fi
    cmake --build . --config Release -j$(nproc)
    if [ $? -ne 0 ]; then
        echo "Error: ZXing build failed"
        exit 1
    fi
    sudo cmake --install . --config Release
    if [ $? -ne 0 ]; then
        echo "Error: ZXing installation failed"
        exit 1
    fi
    sudo ldconfig
    cd ../..
    echo "ZXing library installed successfully"
else
    echo "[build.sh] ZXing found in system, will use system package."
fi

# 创建构建目录
mkdir -p build
cd build

# 配置 CMake
echo "Configuring CMake..."
cmake .. -DCMAKE_INSTALL_PREFIX=.. -DCMAKE_BUILD_TYPE=Release
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

cd ..

echo "Building Go library..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    go build -o bin/libzxing.dylib -buildmode=c-shared zxing.go
else
    go build -o bin/libzxing.so -buildmode=c-shared zxing.go
fi
if [ $? -ne 0 ]; then
    echo "Error: Go build failed"
    exit 1
fi

echo "Build completed successfully!" 