# ZXing WASM 演示构建脚本

param(
    [string]$Action = "help"
)

function Show-Help {
    Write-Host "ZXing WASM 构建和测试脚本" -ForegroundColor Green
    Write-Host ""
    Write-Host "用法: .\build_wasm_demo.ps1 [action]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "可用操作:" -ForegroundColor Cyan
    Write-Host "  build    - 构建 Go 包"
    Write-Host "  test     - 运行测试"
    Write-Host "  demo     - 运行演示程序"
    Write-Host "  wasm     - 构建 WASM 模块（需要 Emscripten）"
    Write-Host "  server   - 启动测试服务器"
    Write-Host "  all      - 执行所有操作"
    Write-Host "  help     - 显示此帮助信息"
}

function Test-Go {
    Write-Host "检查 Go 环境..." -ForegroundColor Yellow
    
    try {
        $goVersion = go version
        Write-Host "✓ Go 已安装: $goVersion" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "✗ Go 未安装或不在 PATH 中" -ForegroundColor Red
        return $false
    }
}

function Build-GoPackages {
    Write-Host "构建 Go 包..." -ForegroundColor Yellow
    
    Write-Host "  构建 zxing 包..."
    go build ./pkg/zxing/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ zxing 包构建失败" -ForegroundColor Red
        return $false
    }
    
    Write-Host "  构建示例程序..."
    go build ./cmd/wasm-example/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ 示例程序构建失败" -ForegroundColor Red
        return $false
    }
    
    Write-Host "✓ Go 包构建成功" -ForegroundColor Green
    return $true
}

function Run-Tests {
    Write-Host "运行测试..." -ForegroundColor Yellow
    
    go test ./pkg/zxing/ -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ 测试失败" -ForegroundColor Red
        return $false
    }
    
    Write-Host "✓ 所有测试通过" -ForegroundColor Green
    return $true
}

function Run-Demo {
    Write-Host "运行演示程序..." -ForegroundColor Yellow
    
    go run ./cmd/wasm-example/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ 演示程序运行失败" -ForegroundColor Red
        return $false
    }
    
    Write-Host "✓ 演示程序运行完成" -ForegroundColor Green
    return $true
}

function Build-WASM {
    Write-Host "构建 WASM 模块..." -ForegroundColor Yellow
    
    # 检查 Emscripten
    try {
        $emccVersion = emcc --version
        Write-Host "✓ Emscripten 已安装" -ForegroundColor Green
    } catch {
        Write-Host "✗ Emscripten 未安装，跳过 WASM 构建" -ForegroundColor Yellow
        Write-Host "  安装指南: https://emscripten.org/docs/getting_started/downloads.html"
        return $false
    }
    
    # 执行 WASM 构建
    Set-Location wasm
    try {
        .\build_simple.bat
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ WASM 模块构建成功" -ForegroundColor Green
            return $true
        } else {
            Write-Host "✗ WASM 模块构建失败" -ForegroundColor Red
            return $false
        }
    } finally {
        Set-Location ..
    }
}

function Start-TestServer {
    Write-Host "启动测试服务器..." -ForegroundColor Yellow
    
    # 检查是否有 Python
    try {
        python --version | Out-Null
        Write-Host "使用 Python 启动服务器..."
        Write-Host "访问 http://localhost:8000/wasm/test_simple.html 查看测试页面" -ForegroundColor Cyan
        python -m http.server 8000
    } catch {
        try {
            # 尝试使用 PowerShell 的简单服务器
            Write-Host "使用 PowerShell 启动简单服务器..."
            Write-Host "访问 http://localhost:8080/wasm/test_simple.html 查看测试页面" -ForegroundColor Cyan
            
            # 这里可以添加 PowerShell HTTP 服务器代码
            Write-Host "请手动打开 wasm/test_simple.html 文件进行测试" -ForegroundColor Yellow
        } catch {
            Write-Host "✗ 无法启动服务器，请手动打开 wasm/test_simple.html" -ForegroundColor Red
        }
    }
}

# 主逻辑
switch ($Action.ToLower()) {
    "build" {
        if (Test-Go) {
            Build-GoPackages
        }
    }
    
    "test" {
        if (Test-Go) {
            Run-Tests
        }
    }
    
    "demo" {
        if (Test-Go) {
            Run-Demo
        }
    }
    
    "wasm" {
        Build-WASM
    }
    
    "server" {
        Start-TestServer
    }
    
    "all" {
        Write-Host "执行完整构建和测试流程..." -ForegroundColor Green
        Write-Host ""
        
        if (-not (Test-Go)) {
            exit 1
        }
        
        if (-not (Build-GoPackages)) {
            exit 1
        }
        
        if (-not (Run-Tests)) {
            exit 1
        }
        
        Run-Demo
        
        Write-Host ""
        Write-Host "可选: 构建 WASM 模块 (需要 Emscripten)"
        Build-WASM
        
        Write-Host ""
        Write-Host "✓ 所有操作完成！" -ForegroundColor Green
        Write-Host "下一步: 运行 '.\build_wasm_demo.ps1 server' 启动测试服务器" -ForegroundColor Cyan
    }
    
    default {
        Show-Help
    }
}