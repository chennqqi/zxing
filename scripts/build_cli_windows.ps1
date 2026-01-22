# Windows CLI构建脚本
# 编译有用的CLI工具

param(
    [string]$BuildType = "Release"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building CLI Tools for Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 获取脚本所在目录的父目录（项目根目录）
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

# 创建输出目录
$BinDir = "bin\windows"
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
}

$SuccessCount = 0
$FailCount = 0

# 要编译的CLI工具列表
$CliTools = @(
    @{
        Name = "zxing-cli"
        Path = "cmd\zxing-cli"
        Description = "Main CLI tool for barcode/QR code scanning"
    },
    @{
        Name = "wasm-example"
        Path = "cmd\wasm-example"
        Description = "WASM example application"
    }
    # Note: zxing-server uses old zxing package and requires CGO
    # It will be built separately when CGO is available
)

Write-Host "Building CLI tools..." -ForegroundColor Cyan
Write-Host ""

foreach ($tool in $CliTools) {
    Write-Host "Building $($tool.Name)..." -ForegroundColor Yellow
    Write-Host "  Path: $($tool.Path)" -ForegroundColor Gray
    Write-Host "  Description: $($tool.Description)" -ForegroundColor Gray
    
    $outputPath = Join-Path $BinDir "$($tool.Name).exe"
    
    # 构建命令
    $buildArgs = @(
        "build",
        "-o", $outputPath,
        "./$($tool.Path)"
    )
    
    # 注意：CGO编译需要gcc编译器
    # 如果gcc不可用，可以设置CGO_ENABLED=0，但功能会受限
    # 暂时不强制启用CGO，让Go自动检测
    # if ($tool.Name -eq "zxing-server" -or $tool.Name -eq "zxing-cli") {
    #     $env:CGO_ENABLED = "1"
    # }
    
    try {
        $output = & go @buildArgs 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  Build output:" -ForegroundColor Gray
            $output | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        }
        
        if ($LASTEXITCODE -eq 0) {
            if (Test-Path $outputPath) {
                $size = (Get-Item $outputPath).Length / 1KB
                Write-Host "  ✅ Success: $outputPath ($([math]::Round($size, 2)) KB)" -ForegroundColor Green
                $SuccessCount++
            } else {
                Write-Host "  ❌ Failed: Output file not found" -ForegroundColor Red
                $FailCount++
            }
        } else {
            Write-Host "  ❌ Failed: Build error (exit code $LASTEXITCODE)" -ForegroundColor Red
            $FailCount++
        }
    } catch {
        Write-Host "  ❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
        $FailCount++
    }
    
    Write-Host ""
}

# 恢复环境变量
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

# 总结
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Success: $SuccessCount" -ForegroundColor Green
Write-Host "Failed: $FailCount" -ForegroundColor $(if ($FailCount -gt 0) { "Red" } else { "Green" })
Write-Host ""

if ($FailCount -eq 0) {
    Write-Host "All CLI tools built successfully!" -ForegroundColor Green
    Write-Host "Output directory: $BinDir" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Available tools:" -ForegroundColor Yellow
    Get-ChildItem $BinDir -Filter "*.exe" | ForEach-Object {
        $size = $_.Length / 1KB
        Write-Host "  - $($_.Name) ($([math]::Round($size, 2)) KB)" -ForegroundColor White
    }
    exit 0
} else {
    Write-Host "Some builds failed. Please check the errors above." -ForegroundColor Red
    exit 1
}
