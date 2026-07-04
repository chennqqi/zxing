# WASM构建并保存脚本 (Windows)
# 编译ZXingCPP为WASM，并保存到wasm目录

param(
    [string]$BuildType = "Release"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building WASM Module for Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Type: $BuildType" -ForegroundColor Yellow
Write-Host ""

# 检查Emscripten SDK
$emsdkPath = $env:EMSDK
if (-not $emsdkPath) {
    Write-Host "Error: EMSDK environment variable not set" -ForegroundColor Red
    Write-Host "Please install Emscripten SDK first:" -ForegroundColor Yellow
    Write-Host "  https://emscripten.org/docs/getting_started/downloads.html" -ForegroundColor Yellow
    exit 1
}

Write-Host "Using Emscripten SDK at: $emsdkPath" -ForegroundColor Cyan

# 检查emcc命令
if (-not (Get-Command emcc -ErrorAction SilentlyContinue)) {
    Write-Host "Error: emcc command not found" -ForegroundColor Red
    Write-Host "Please activate Emscripten SDK first:" -ForegroundColor Yellow
    Write-Host "  source $emsdkPath\emsdk_env.ps1" -ForegroundColor Yellow
    exit 1
}

# 获取脚本所在目录的父目录（项目根目录）
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

# Initialize zxing-cpp submodule
Write-Host "Initializing zxing-cpp submodule..." -ForegroundColor Cyan
git submodule update --init --recursive
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to initialize zxing-cpp submodule" -ForegroundColor Red
    exit 1
}

# 创建构建目录
$BuildDir = "build-wasm"
if (Test-Path $BuildDir) {
    Write-Host "Removing existing build directory..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $BuildDir
}
New-Item -ItemType Directory -Path $BuildDir | Out-Null

Push-Location $BuildDir

try {
    # 配置CMake - 使用WASM工具链
    Write-Host "Configuring CMake for WASM build..." -ForegroundColor Cyan
    $toolchainFile = Join-Path $emsdkPath "upstream\emscripten\cmake\Modules\Platform\Emscripten.cmake"
    
    if (-not (Test-Path $toolchainFile)) {
        Write-Host "Error: Emscripten toolchain file not found: $toolchainFile" -ForegroundColor Red
        exit 1
    }
    
    $cmakeArgs = @(
        "-G", "MinGW Makefiles",
        "-DCMAKE_BUILD_TYPE=$BuildType",
        "-DCMAKE_TOOLCHAIN_FILE=$toolchainFile",
        "-DBUILD_SHARED_LIBS=OFF",
        "-DZXING_EXAMPLES=OFF",
        "-DZXING_BLACKBOX_TESTS=OFF",
        "-DZXING_UNIT_TESTS=OFF",
        "-DZXING_INSTALL_TEST=OFF",
        "-DZXING_PYTHON_MODULE=OFF",
        "-Wno-dev",
        ".."
    )
    
    # 使用CMakeLists-wasm.txt作为配置
    Copy-Item "..\CMakeLists-wasm.txt" "..\CMakeLists.txt.backup" -ErrorAction SilentlyContinue
    Copy-Item "..\CMakeLists-wasm.txt" "..\CMakeLists.txt" -Force
    
    & cmake @cmakeArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: CMake configuration failed" -ForegroundColor Red
        exit 1
    }
    
    # 恢复原始CMakeLists.txt
    if (Test-Path "..\CMakeLists.txt.backup") {
        Move-Item "..\CMakeLists.txt.backup" "..\CMakeLists.txt" -Force
    }
    
    # 构建
    Write-Host "Building WASM module..." -ForegroundColor Cyan
    & cmake --build . --config $BuildType
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Build failed" -ForegroundColor Red
        exit 1
    }
    
    # 创建wasm目录
    $WasmDir = Join-Path $ProjectRoot "wasm"
    if (-not (Test-Path $WasmDir)) {
        New-Item -ItemType Directory -Path $WasmDir -Force | Out-Null
    }
    
    # 复制WASM文件
    Write-Host "Copying WASM files..." -ForegroundColor Cyan
    
    # 查找WASM文件
    $WasmFile = Get-ChildItem -Path ".\bin" -Filter "*.wasm" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($WasmFile) {
        Copy-Item $WasmFile.FullName (Join-Path $WasmDir "zxing.wasm") -Force
        Write-Host "  Copied: zxing.wasm" -ForegroundColor Green
    } else {
        Write-Host "  Warning: WASM file not found" -ForegroundColor Yellow
    }
    
    # 查找JS文件
    $JsFile = Get-ChildItem -Path ".\bin" -Filter "*.js" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($JsFile) {
        Copy-Item $JsFile.FullName (Join-Path $WasmDir "zxing.js") -Force
        Write-Host "  Copied: zxing.js" -ForegroundColor Green
    } else {
        Write-Host "  Warning: JS file not found" -ForegroundColor Yellow
    }
    
    # 验证文件
    Write-Host ""
    Write-Host "Verifying WASM files..." -ForegroundColor Cyan
    $WasmFilePath = Join-Path $WasmDir "zxing.wasm"
    $JsFilePath = Join-Path $WasmDir "zxing.js"
    
    if (Test-Path $WasmFilePath) {
        $size = (Get-Item $WasmFilePath).Length / 1MB
        Write-Host "  zxing.wasm: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  zxing.wasm: NOT FOUND" -ForegroundColor Red
    }
    
    if (Test-Path $JsFilePath) {
        $size = (Get-Item $JsFilePath).Length / 1KB
        Write-Host "  zxing.js: $([math]::Round($size, 2)) KB" -ForegroundColor Green
    } else {
        Write-Host "  zxing.js: NOT FOUND" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "WASM build completed successfully!" -ForegroundColor Green
    Write-Host "Files saved to: $WasmDir" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}
