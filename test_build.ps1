# 简单的构建测试脚本

Write-Host "=== ZXing WASM 项目构建测试 ===" -ForegroundColor Green
Write-Host ""

# 测试 Go 环境
Write-Host "1. 检查 Go 环境..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "   ✓ $goVersion" -ForegroundColor Green
} catch {
    Write-Host "   ✗ Go 未安装" -ForegroundColor Red
    exit 1
}

# 构建 Go 包
Write-Host ""
Write-Host "2. 构建 Go 包..." -ForegroundColor Yellow

Write-Host "   构建 zxing 包..."
go build ./pkg/zxing/
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ zxing 包构建成功" -ForegroundColor Green
} else {
    Write-Host "   ✗ zxing 包构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "   构建示例程序..."
go build ./cmd/wasm-example/
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ 示例程序构建成功" -ForegroundColor Green
} else {
    Write-Host "   ✗ 示例程序构建失败" -ForegroundColor Red
    exit 1
}

# 运行测试
Write-Host ""
Write-Host "3. 运行测试..." -ForegroundColor Yellow
go test ./pkg/zxing/ -v
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ 所有测试通过" -ForegroundColor Green
} else {
    Write-Host "   ✗ 测试失败" -ForegroundColor Red
    exit 1
}

# 运行演示
Write-Host ""
Write-Host "4. 运行演示程序..." -ForegroundColor Yellow
go run ./cmd/wasm-example/
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ 演示程序运行成功" -ForegroundColor Green
} else {
    Write-Host "   ✗ 演示程序运行失败" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 构建测试完成 ===" -ForegroundColor Green
Write-Host ""
Write-Host "下一步操作:" -ForegroundColor Cyan
Write-Host "1. 安装 Emscripten SDK 来构建真实的 WASM 模块"
Write-Host "2. 运行 'cd wasm && build_simple.bat' 构建 WASM"
Write-Host "3. 打开 wasm/test_simple.html 测试 WASM 功能"