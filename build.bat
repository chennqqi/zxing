@echo off
setlocal enabledelayedexpansion

:: 检查是否安装了 CMake
where cmake >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: CMake is not installed
    exit /b 1
)

:: 检查是否安装了 Go
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed
    exit /b 1
)

:: 创建构建目录
if not exist build mkdir build
cd build

:: 配置 CMake
echo Configuring CMake...
cmake .. -G "Visual Studio 17 2022" -A x64 ^
    -DCMAKE_INSTALL_PREFIX=.. ^
    -DCMAKE_BUILD_TYPE=Release

if %ERRORLEVEL% neq 0 (
    echo Error: CMake configuration failed
    exit /b 1
)

:: 构建
echo Building...
cmake --build . --config Release

if %ERRORLEVEL% neq 0 (
    echo Error: Build failed
    exit /b 1
)

:: 安装
echo Installing...
cmake --install . --config Release

if %ERRORLEVEL% neq 0 (
    echo Error: Installation failed
    exit /b 1
)

:: 返回上级目录
cd ..

:: 构建 Go 库
echo Building Go library...
go build -o bin/zxing.dll -buildmode=c-shared

if %ERRORLEVEL% neq 0 (
    echo Error: Go build failed
    exit /b 1
)

echo Build completed successfully! 