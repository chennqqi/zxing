# ZXing 集成测试脚本
# 测试 CGO 和 WASM 两种后端

Write-Host "=== ZXing 项目集成测试 ===" -ForegroundColor Green
Write-Host ""

# 1. 检查 Go 环境
Write-Host "1. 检查 Go 环境..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go 环境检查失败" -ForegroundColor Red
    exit 1
}

# 2. 清理并重新构建
Write-Host "`n2. 清理并重新构建..." -ForegroundColor Yellow
go clean -cache
go mod tidy

# 3. 构建各个包
Write-Host "`n3. 构建各个包..." -ForegroundColor Yellow

Write-Host "构建 pkg/zxing..."
go build ./pkg/zxing/
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ pkg/zxing 构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ pkg/zxing 构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "构建 pkg/wasm..."
go build ./pkg/wasm/
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ pkg/wasm 构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ pkg/wasm 构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "构建 cmd/wasm-example..."
go build ./cmd/wasm-example/
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ cmd/wasm-example 构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ cmd/wasm-example 构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "构建 cmd/test_backends..."
go build ./cmd/test_backends/
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ cmd/test_backends 构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ cmd/test_backends 构建失败" -ForegroundColor Red
    exit 1
}

# 4. 运行单元测试
Write-Host "`n4. 运行单元测试..." -ForegroundColor Yellow
go test ./pkg/zxing -v
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 单元测试通过" -ForegroundColor Green
} else {
    Write-Host "❌ 单元测试失败" -ForegroundColor Red
    exit 1
}

# 5. 测试后端切换
Write-Host "`n5. 测试后端切换..." -ForegroundColor Yellow
go run ./cmd/test_backends
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 后端切换测试通过" -ForegroundColor Green
} else {
    Write-Host "❌ 后端切换测试失败" -ForegroundColor Red
    exit 1
}

# 6. 测试 WASM 示例
Write-Host "`n6. 测试 WASM 示例..." -ForegroundColor Yellow
go run ./cmd/wasm-example
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ WASM 示例测试通过" -ForegroundColor Green
} else {
    Write-Host "❌ WASM 示例测试失败" -ForegroundColor Red
    exit 1
}

# 7. 测试 CGO 后端
Write-Host "`n7. 测试 CGO 后端..." -ForegroundColor Yellow
$env:ZXING_BACKEND="cgo"
go run ./cmd/wasm-example
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ CGO 后端测试通过" -ForegroundColor Green
} else {
    Write-Host "❌ CGO 后端测试失败" -ForegroundColor Red
    exit 1
}

# 8. 性能测试
Write-Host "`n8. 运行性能测试..." -ForegroundColor Yellow
go test ./pkg/zxing -bench=. -run=^$
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 性能测试完成" -ForegroundColor Green
} else {
    Write-Host "⚠️ 性能测试部分失败（可能是正常的）" -ForegroundColor Yellow
}

Write-Host "`n=== 集成测试完成 ===" -ForegroundColor Green
Write-Host "✅ 所有核心功能测试通过" -ForegroundColor Green
Write-Host ""
Write-Host "项目状态总结:" -ForegroundColor Cyan
Write-Host "- CGO 后端: 正常工作" -ForegroundColor Green
Write-Host "- WASM 后端: 正常工作（模拟模式）" -ForegroundColor Green
Write-Host "- 统一接口: 正常工作" -ForegroundColor Green
Write-Host "- 后端切换: 正常工作" -ForegroundColor Green
Write-Host ""
Write-Host "下一步建议:" -ForegroundColor Yellow
Write-Host "1. 如需真实 WASM 模块，请安装 Emscripten 并编译 zxing C++ 库"
Write-Host "2. 在浏览器中测试 WASM 功能"
Write-Host "3. 进行更深入的性能测试"
