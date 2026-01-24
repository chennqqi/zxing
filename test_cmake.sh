#!/bin/bash

echo "Testing CMake configuration..."

# 清理构建目录
rm -rf build
mkdir -p build
cd build

# 配置 CMake
echo "Configuring CMake..."
cmake .. -DCMAKE_INSTALL_PREFIX=.. -DCMAKE_BUILD_TYPE=Release

if [ $? -eq 0 ]; then
    echo "CMake configuration successful"
    echo "Attempting to build..."
    cmake --build . --config Release
    if [ $? -eq 0 ]; then
        echo "Build successful!"
    else
        echo "Build failed"
        exit 1
    fi
else
    echo "CMake configuration failed"
    exit 1
fi 