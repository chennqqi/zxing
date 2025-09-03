# ZXing WASM 完整构建脚本

param(
    [string]$Target = "all"
)

Write-Host "ZXing WASM 项目构建脚本" -ForegroundColor Green
Write-Host "=========================" -ForegroundColor Green

function Test-Go {
    Write-Host "`n检查 Go 环境..." -ForegroundColor Yellow
    
    try {
        $goVersion = go version
        Write-Host "✅ $goVersion" -ForegroundColor Green
    } catch {
        Write-Host "❌ 未找到 Go 编译器" -ForegroundColor Red
        exit 1
    }
}

function Build-GoPackages {
    Write-Host "`n构建 Go 包..." -ForegroundColor Yellow
    
    Write-Host "构建 pkg/zxing..."
    go build ./pkg/zxing/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ pkg/zxing 构建失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "构建 pkg/wasm..."
    go build ./pkg/wasm/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ pkg/wasm 构建失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "构建 cmd/wasm-example..."
    go build ./cmd/wasm-example/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ cmd/wasm-example 构建失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ Go 包构建成功" -ForegroundColor Green
}

function Run-Tests {
    Write-Host "`n运行测试..." -ForegroundColor Yellow
    
    go test ./pkg/zxing/ -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 测试失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ 所有测试通过" -ForegroundColor Green
}

function Build-WASMDemo {
    Write-Host "`n构建 WASM 演示文件..." -ForegroundColor Yellow
    
    Push-Location wasm
    
    if (Test-Path "build_demo.sh") {
        if (Get-Command "bash" -ErrorAction SilentlyContinue) {
            bash build_demo.sh
        } else {
            Write-Host "创建演示文件..."
            
            # 创建 JavaScript 模拟器
            @"
// ZXing WASM 模拟器 - 用于演示和测试
function ZXingWASM() {
    return new Promise((resolve) => {
        setTimeout(() => {
            const module = {
                encode_text_to_qr: function(text, width, height) {
                    console.log(`模拟编码: "${text}" ${width}x${height}`);
                    const data = new Array(width * height);
                    for (let i = 0; i < data.length; i++) {
                        data[i] = (i % 20 < 10) ? 0 : 255;
                    }
                    return {
                        success: true, width: width, height: height,
                        data: { size: () => data.length, get: (i) => data[i] },
                        error_code: 0, error_message: ""
                    };
                },
                decode_image_data: function(dataPtr, width, height, channels) {
                    return {
                        success: true, text: "Demo: Hello from WASM!",
                        format: "QR_CODE", error_code: 0, error_message: ""
                    };
                },
                _malloc: function(size) { return new ArrayBuffer(size); },
                _free: function(ptr) { },
                HEAPU8: { set: function(data, offset) { } }
            };
            resolve(module);
        }, 500);
    });
}
if (typeof window !== 'undefined') { window.ZXingWASM = ZXingWASM; }
"@ | Out-File -FilePath "zxing.js" -Encoding UTF8
            
            # 创建空的 WASM 文件
            "" | Out-File -FilePath "zxing.wasm" -Encoding UTF8
        }
        
        Write-Host "✅ WASM 演示文件创建成功" -ForegroundColor Green
    }
    
    Pop-Location
}

function Run-Example {
    Write-Host "`n运行示例程序..." -ForegroundColor Yellow
    
    go run ./cmd/wasm-example/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 示例程序运行失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ 示例程序运行成功" -ForegroundColor Green
}

function Show-Summary {
    Write-Host "`n构建完成总结:" -ForegroundColor Green
    Write-Host "===============" -ForegroundColor Green
    Write-Host "✅ Go 包编译成功"
    Write-Host "✅ 单元测试通过"
    Write-Host "✅ 示例程序运行正常"
    Write-Host "✅ WASM 演示文件已创建"
    Write-Host ""
    Write-Host "下一步操作:" -ForegroundColor Yellow
    Write-Host "1. 在浏览器中打开 wasm/test.html 测试 WASM 功能"
    Write-Host "2. 运行 'go run ./cmd/wasm-example/' 测试 Go 程序"
    Write-Host "3. 如需真实 WASM 模块，请安装 Emscripten 并运行 wasm/build.sh"
    Write-Host ""
    Write-Host "项目结构:" -ForegroundColor Cyan
    Write-Host "├── pkg/zxing/     # 统一接口层"
    Write-Host "├── pkg/wasm/      # WASM 运行时"
    Write-Host "├── cmd/wasm-example/ # 示例程序"
    Write-Host "├── wasm/          # WASM 构建文件"
    Write-Host "└── doc/           # 文档"
}

# 主执行逻辑
switch ($Target.ToLower()) {
    "all" {
        Test-Go
        Build-GoPackages
        Run-Tests
        Build-WASMDemo
        Run-Example
        Show-Summary
    }
    "build" {
        Test-Go
        Build-GoPackages
    }
    "test" {
        Run-Tests
    }
    "wasm" {
        Build-WASMDemo
    }
    "example" {
        Run-Example
    }
    default {
        Write-Host "用法: .\build_all.ps1 [all|build|test|wasm|example]"
        exit 1
    }
}