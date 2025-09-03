# WASM构建脚本 for Windows
# 需要先安装Emscripten SDK

param(
    [string]$BuildType = "Release",
    [string]$BuildDir = "build-wasm"
)

Write-Host "Building ZXing WASM module..." -ForegroundColor Green

# 检查Emscripten是否安装
$emsdkPath = $env:EMSDK
if (-not $emsdkPath) {
    Write-Host "Error: EMSDK environment variable not set. Please install Emscripten SDK first." -ForegroundColor Red
    Write-Host "Installation guide: https://emscripten.org/docs/getting_started/downloads.html" -ForegroundColor Yellow
    exit 1
}

Write-Host "Using Emscripten SDK at: $emsdkPath" -ForegroundColor Cyan

# 创建构建目录
if (Test-Path $BuildDir) {
    Write-Host "Removing existing build directory..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $BuildDir
}
New-Item -ItemType Directory -Path $BuildDir | Out-Null

# 进入构建目录
Push-Location $BuildDir

try {
    # 配置CMake
    Write-Host "Configuring CMake..." -ForegroundColor Cyan
    $cmakeArgs = @(
        "-G", "MinGW Makefiles",
        "-DCMAKE_BUILD_TYPE=$BuildType",
        "-DCMAKE_TOOLCHAIN_FILE=$emsdkPath\upstream\emscripten\cmake\Modules\Platform\Emscripten.cmake",
        ".."
    )
    
    $cmakeResult = & cmake @cmakeArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Host "CMake configuration failed!" -ForegroundColor Red
        exit 1
    }
    
    # 构建项目
    Write-Host "Building project..." -ForegroundColor Cyan
    $buildResult = & cmake --build . --config $BuildType
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
    
    # 检查输出文件
    $wasmFile = Join-Path $BuildDir "bin\zxingwrapper.wasm"
    $jsFile = Join-Path $BuildDir "bin\zxingwrapper.js"
    
    if (Test-Path $wasmFile) {
        $wasmSize = (Get-Item $wasmFile).Length
        Write-Host "WASM file generated: $wasmFile ($([math]::Round($wasmSize/1MB, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "Warning: WASM file not found!" -ForegroundColor Yellow
    }
    
    if (Test-Path $jsFile) {
        $jsSize = (Get-Item $jsFile).Length
        Write-Host "JS file generated: $jsFile ($([math]::Round($jsSize/1KB, 2)) KB)" -ForegroundColor Green
    } else {
        Write-Host "Warning: JS file not found!" -ForegroundColor Yellow
    }
    
    # 复制文件到项目目录
    Write-Host "Copying files to project directory..." -ForegroundColor Cyan
    $wasmDir = "..\wasm"
    if (-not (Test-Path $wasmDir)) {
        New-Item -ItemType Directory -Path $wasmDir | Out-Null
    }
    
    if (Test-Path $wasmFile) {
        Copy-Item $wasmFile $wasmDir -Force
        Write-Host "Copied WASM file to: $wasmDir" -ForegroundColor Green
    }
    
    if (Test-Path $jsFile) {
        Copy-Item $jsFile $wasmDir -Force
        Write-Host "Copied JS file to: $wasmDir" -ForegroundColor Green
    }
    
    Write-Host "WASM build completed successfully!" -ForegroundColor Green
    
} catch {
    Write-Host "Build failed with error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} finally {
    # 返回原目录
    Pop-Location
}
