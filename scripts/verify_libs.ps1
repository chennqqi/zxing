# 验证lib目录中的静态库文件

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Verifying Library Files" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 获取脚本所在目录的父目录（项目根目录）
$ProjectRoot = Split-Path -Parent $PSScriptRoot

$AllFound = $true

# 检查Windows静态库
Write-Host "Checking Windows static libraries..." -ForegroundColor Cyan
$WindowsLibDir = Join-Path $ProjectRoot "lib\windows\x64"

if (Test-Path $WindowsLibDir) {
    $ZXingLib = Join-Path $WindowsLibDir "ZXing.lib"
    $WrapperLib = Join-Path $WindowsLibDir "zxingwrapper.lib"
    
    if (Test-Path $ZXingLib) {
        $size = (Get-Item $ZXingLib).Length / 1MB
        Write-Host "  [OK] ZXing.lib: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] ZXing.lib" -ForegroundColor Red
        $AllFound = $false
    }
    
    if (Test-Path $WrapperLib) {
        $size = (Get-Item $WrapperLib).Length / 1KB
        Write-Host "  [OK] zxingwrapper.lib: $([math]::Round($size, 2)) KB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] zxingwrapper.lib" -ForegroundColor Red
        $AllFound = $false
    }
} else {
    Write-Host "  [MISSING] Directory: $WindowsLibDir" -ForegroundColor Red
    $AllFound = $false
}

Write-Host ""

# 检查Linux静态库
Write-Host "Checking Linux static libraries..." -ForegroundColor Cyan
$LinuxLibDir = Join-Path $ProjectRoot "lib\linux\x64"

if (Test-Path $LinuxLibDir) {
    $ZXingLib = Join-Path $LinuxLibDir "libZXing.a"
    $WrapperLib = Join-Path $LinuxLibDir "libzxingwrapper.a"
    
    if (Test-Path $ZXingLib) {
        $size = (Get-Item $ZXingLib).Length / 1MB
        Write-Host "  [OK] libZXing.a: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] libZXing.a" -ForegroundColor Red
        $AllFound = $false
    }
    
    if (Test-Path $WrapperLib) {
        $size = (Get-Item $WrapperLib).Length / 1KB
        Write-Host "  [OK] libzxingwrapper.a: $([math]::Round($size, 2)) KB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] libzxingwrapper.a" -ForegroundColor Red
        $AllFound = $false
    }
} else {
    Write-Host "  [MISSING] Directory: $LinuxLibDir" -ForegroundColor Red
    $AllFound = $false
}

Write-Host ""

# 检查WASM文件
Write-Host "Checking WASM files..." -ForegroundColor Cyan
$WasmDir = Join-Path $ProjectRoot "wasm"

if (Test-Path $WasmDir) {
    $WasmFile = Join-Path $WasmDir "zxing.wasm"
    $JsFile = Join-Path $WasmDir "zxing.js"
    
    if (Test-Path $WasmFile) {
        $size = (Get-Item $WasmFile).Length / 1MB
        Write-Host "  [OK] zxing.wasm: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] zxing.wasm" -ForegroundColor Red
        $AllFound = $false
    }
    
    if (Test-Path $JsFile) {
        $size = (Get-Item $JsFile).Length / 1KB
        Write-Host "  [OK] zxing.js: $([math]::Round($size, 2)) KB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] zxing.js" -ForegroundColor Red
        $AllFound = $false
    }
} else {
    Write-Host "  [MISSING] Directory: $WasmDir" -ForegroundColor Red
    $AllFound = $false
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
if ($AllFound) {
    Write-Host "All library files are present!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Some library files are missing!" -ForegroundColor Red
    Write-Host "Please run the build scripts to generate them." -ForegroundColor Yellow
    exit 1
}
