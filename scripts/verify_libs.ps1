# Verify static library files in the lib/ directory.
#
# Checks each platform directory lib/{os}-{arch}/ for libZXing.a and
# libzxingwrapper.a. Missing platform directories are reported but do not
# fail the check — a directory that EXISTS but has missing files does fail.

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Verifying Library Files" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ProjectRoot = Split-Path -Parent $PSScriptRoot

$AllFound = $true
$FoundAny = $false

function Test-Platform {
    param([string]$Name)

    $dir = Join-Path $ProjectRoot "lib\$Name"

    Write-Host "Checking $Name..." -ForegroundColor Cyan

    if (-not (Test-Path $dir)) {
        Write-Host "  [SKIP] Directory not present: lib\$Name" -ForegroundColor DarkGray
        Write-Host ""
        return
    }

    $script:FoundAny = $true
    foreach ($lib in @("libZXing.a", "libzxingwrapper.a")) {
        $path = Join-Path $dir $lib
        if (Test-Path $path) {
            $size = (Get-Item $path).Length / 1KB
            Write-Host "  [OK] ${lib}: $([math]::Round($size, 0)) KB" -ForegroundColor Green
        } else {
            Write-Host "  [MISSING] $lib" -ForegroundColor Red
            $script:AllFound = $false
        }
    }
    Write-Host ""
}

Test-Platform "linux-x64"
Test-Platform "linux-arm64"
Test-Platform "windows-x64"
Test-Platform "windows-arm64"
Test-Platform "darwin-x64"
Test-Platform "darwin-arm64"

# Check WASM files
Write-Host "Checking WASM files..." -ForegroundColor Cyan
$WasmDir = Join-Path $ProjectRoot "wasm"

if (Test-Path $WasmDir) {
    $WasmFile = Join-Path $WasmDir "zxingwrapper.wasm"
    if (Test-Path $WasmFile) {
        $size = (Get-Item $WasmFile).Length / 1MB
        Write-Host "  [OK] zxingwrapper.wasm: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] zxingwrapper.wasm" -ForegroundColor Red
        $AllFound = $false
    }
} else {
    Write-Host "  [MISSING] Directory: $WasmDir" -ForegroundColor Red
    $AllFound = $false
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
if ($AllFound) {
    Write-Host "All present library files verified!" -ForegroundColor Green
    if (-not $FoundAny) {
        Write-Host "Note: no platform lib/ directories found." -ForegroundColor Yellow
    }
    exit 0
} else {
    Write-Host "Some library files are missing!" -ForegroundColor Red
    Write-Host "Please run the build scripts to generate them." -ForegroundColor Yellow
    exit 1
}
