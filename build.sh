#!/bin/bash

# 检查并安装 ZXing 库（可选）
install_zxing() {
    echo "Checking for ZXing library..."
    
    # 检查不同的 pkg-config 包名
    local pkg_names=("zxing-cpp" "zxing" "libzxing")
    local found=false
    
    for pkg in "${pkg_names[@]}"; do
        if pkg-config --exists "$pkg" 2>/dev/null; then
            echo "ZXing library found via pkg-config: $pkg"
            found=true
            break
        fi
    done
    
    if [ "$found" = true ]; then
        echo "ZXing library already installed"
        return 0
    fi
    
    # 检查是否可以通过 CMake 找到
    if cmake --find-package -DNAME=ZXing -DCOMPILER_ID=GNU -DLANGUAGE=CXX -DMODE=EXIST 2>/dev/null; then
        echo "ZXing library found via CMake"
        return 0
    fi
    
    echo "ZXing library not found, installing from source..."
    
    # 从源码编译安装 ZXing
    if [ ! -d "zxing-cpp" ]; then
        git clone https://github.com/zxing-cpp/zxing-cpp.git --recursive --single-branch --depth 1
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
}

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

# 检查是否需要手动安装 ZXing（可选）
# 可以通过环境变量 ZXING_INSTALL=1 来强制安装
if [ "${ZXING_INSTALL:-0}" = "1" ]; then
    install_zxing
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
    echo "If this is due to missing ZXing library, try: ZXING_INSTALL=1 ./build.sh"
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