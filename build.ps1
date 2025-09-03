# ZXing Go WASM 集成构建脚本 (PowerShell)

param(
    [string]$Target = "build"
)

function Show-Help {
    Write-Host "可用的构建目标:"
    Write-Host "  build        - 构建 Go 包和示例"
    Write-Host "  build-go     - 构建 Go 包"
    Write-Host "  build-example- 构建示例程序"
    Write-Host "  build-wasm   - 构建 WASM 版本"
    Write-Host "  build-cgo    - 构建 CGO 版本"
    Write-Host "  test         - 运行测试"
    Write-Host "  bench        - 运行基准测试"
    Write-Host "  run-example  - 运行示例程序"
    Write-Host "  clean        - 清理构建文件"
    Write-Host "  help         - 显示此帮助信息"
}

function Build-Go {
    Write-Host "构建 Go 包..."
    go build ./pkg/zxing/
    go build ./pkg/wasm/
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Go 包构建失败"
        exit 1
    }
}

function Build-Example {
    Write-Host "构建示例程序..."
    if (!(Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    go build -o bin/wasm-example.exe ./cmd/wasm-example/
    if ($LASTEXITCODE -ne 0) {
        Write-Error "示例程序构建失败"
        exit 1
    }
}

function Build-WASM {
    Write-Host "构建 WASM 版本..."
    if (!(Test-Path "wasm")) {
        New-Item -ItemType Directory -Path "wasm" | Out-Null
    }
    
    $env:GOOS = "js"
    $env:GOARCH = "wasm"
    go build -o wasm/wasm-example.wasm ./cmd/wasm-example/
    
    # 恢复环境变量
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "WASM 版本构建失败"
        exit 1
    }
    
    # 复制 wasm_exec.js
    $goRoot = go env GOROOT
    $wasmExecPath = Join-Path $goRoot "misc\wasm\wasm_exec.js"
    if (Test-Path $wasmExecPath) {
        Copy-Item $wasmExecPath "wasm/"
        Write-Host "已复制 wasm_exec.js"
    }
}

function Build-CGO {
    Write-Host "构建 CGO 版本..."
    if (!(Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    
    $env:CGO_ENABLED = "1"
    go build -tags cgo -o bin/zxing-cgo.exe ./cmd/wasm-example/
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "CGO 版本构建失败"
        exit 1
    }
}

function Run-Tests {
    Write-Host "运行测试..."
    go test ./pkg/zxing/ -v
    go test ./pkg/wasm/ -v
}

function Run-Bench {
    Write-Host "运行基准测试..."
    go test -bench=. ./pkg/zxing/
}

function Run-Example {
    Write-Host "运行示例程序..."
    go run ./cmd/wasm-example/
}

function Clean-Build {
    Write-Host "清理构建文件..."
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    if (Test-Path "wasm") {
        Remove-Item -Recurse -Force "wasm"
    }
}

# 主逻辑
switch ($Target.ToLower()) {
    "build" {
        Build-Go
        Build-Example
    }
    "build-go" {
        Build-Go
    }
    "build-example" {
        Build-Example
    }
    "build-wasm" {
        Build-WASM
    }
    "build-cgo" {
        Build-CGO
    }
    "test" {
        Run-Tests
    }
    "bench" {
        Run-Bench
    }
    "run-example" {
        Run-Example
    }
    "clean" {
        Clean-Build
    }
    "help" {
        Show-Help
    }
    default {
        Write-Host "Unknown target: $Target"
        Show-Help
        exit 1
    }
}