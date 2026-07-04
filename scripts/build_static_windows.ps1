# Windows静态库编译脚本
# 编译ZXingCPP和zxingwrapper为静态库，并保存到lib目录

param(
    [string]$BuildType = "Release",
    [string]$Arch = "x64"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building Static Libraries for Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Type: $BuildType" -ForegroundColor Yellow
Write-Host "Architecture: $Arch" -ForegroundColor Yellow
Write-Host ""

# 检查依赖
Write-Host "Checking dependencies..." -ForegroundColor Cyan
if (-not (Get-Command cmake -ErrorAction SilentlyContinue)) {
    Write-Host "Error: CMake is not installed" -ForegroundColor Red
    exit 1
}

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Git is not installed" -ForegroundColor Red
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
$BuildDir = "build-static-windows"
if (Test-Path $BuildDir) {
    Write-Host "Removing existing build directory..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $BuildDir
}
New-Item -ItemType Directory -Path $BuildDir | Out-Null

Push-Location $BuildDir

try {
    # 配置CMake - 编译静态库
    Write-Host "Configuring CMake for static library build..." -ForegroundColor Cyan
    $cmakeArgs = @(
        "-G", "Visual Studio 17 2022",
        "-A", $Arch,
        "-DCMAKE_BUILD_TYPE=$BuildType",
        "-DBUILD_STATIC_LIB=ON",
        "-DBUILD_SHARED_LIB=OFF",
        "-DCMAKE_INSTALL_PREFIX=$ProjectRoot",
        ".."
    )
    
    & cmake @cmakeArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: CMake configuration failed" -ForegroundColor Red
        exit 1
    }
    
    # 构建
    Write-Host "Building static libraries..." -ForegroundColor Cyan
    & cmake --build . --config $BuildType
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Build failed" -ForegroundColor Red
        exit 1
    }
    
    # 创建lib目录结构
    $LibDir = Join-Path $ProjectRoot "lib\windows\$Arch"
    if (-not (Test-Path $LibDir)) {
        New-Item -ItemType Directory -Path $LibDir -Force | Out-Null
    }
    
    # 复制静态库文件
    Write-Host "Copying static library files..." -ForegroundColor Cyan
    
    # 查找ZXing静态库
    $ZXingLib = Get-ChildItem -Path ".\lib" -Filter "ZXing.lib" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($ZXingLib) {
        Copy-Item $ZXingLib.FullName (Join-Path $LibDir "ZXing.lib") -Force
        Write-Host "  Copied: ZXing.lib" -ForegroundColor Green
    } else {
        Write-Host "  Warning: ZXing.lib not found" -ForegroundColor Yellow
    }
    
    # 查找zxingwrapper静态库
    $WrapperLib = Get-ChildItem -Path ".\lib" -Filter "zxingwrapper.lib" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($WrapperLib) {
        Copy-Item $WrapperLib.FullName (Join-Path $LibDir "zxingwrapper.lib") -Force
        Write-Host "  Copied: zxingwrapper.lib" -ForegroundColor Green
    } else {
        Write-Host "  Warning: zxingwrapper.lib not found" -ForegroundColor Yellow
    }
    
    # 验证文件
    Write-Host ""
    Write-Host "Verifying library files..." -ForegroundColor Cyan
    $ZXingLibPath = Join-Path $LibDir "ZXing.lib"
    $WrapperLibPath = Join-Path $LibDir "zxingwrapper.lib"
    
    if (Test-Path $ZXingLibPath) {
        $size = (Get-Item $ZXingLibPath).Length / 1MB
        Write-Host "  ZXing.lib: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  ZXing.lib: NOT FOUND" -ForegroundColor Red
    }
    
    if (Test-Path $WrapperLibPath) {
        $size = (Get-Item $WrapperLibPath).Length / 1KB
        Write-Host "  zxingwrapper.lib: $([math]::Round($size, 2)) KB" -ForegroundColor Green
    } else {
        Write-Host "  zxingwrapper.lib: NOT FOUND" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Build completed successfully!" -ForegroundColor Green
    Write-Host "Libraries saved to: $LibDir" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}
