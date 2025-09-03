# ZXing WASM 集成指南

## 概述

本指南详细介绍了如何在 ZXing Go wrapper 项目中使用 WebAssembly (WASM) 集成，以避免 CGO 依赖。

## 架构设计

### 核心组件

1. **统一接口层** (`pkg/zxing/interface.go`)
   - 定义了 `ZXing` 接口，支持编码和解码操作
   - 提供统一的 API，无论使用哪种后端

2. **配置管理** (`pkg/zxing/config.go`)
   - 支持环境变量配置
   - 自动后端选择逻辑
   - 调试模式支持

3. **后端实现**
   - **CGO 后端**: 调用现有的 C++ zxing 库
   - **WASM 后端**: 使用 WebAssembly 版本的 zxing
   - **通用实现**: 运行时选择后端

4. **WASM 运行时** (`pkg/wasm/`)
   - 跨平台的 WASM 运行时管理
   - 支持构建标签的条件编译

## 使用方法

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    // 方法1: 使用默认配置（自动选择后端）
    zx, err := zxing.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer zx.Close()

    // 方法2: 明确指定后端
    config := &zxing.Config{
        Backend: zxing.BackendWASM,
        WASMPath: "path/to/zxing.wasm",
        Debug: true,
    }
    zx, err = zxing.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer zx.Close()

    // 编码示例
    opts := &zxing.EncodeOptions{
        Width:  256,
        Height: 256,
        Format: "QR_CODE",
    }
    
    img, err := zx.EncodeText(context.Background(), "Hello, WASM!", opts)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("生成二维码: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
}
```

### 环境变量配置

```bash
# 指定后端类型
export ZXING_BACKEND=wasm  # 或 cgo, auto

# 指定 WASM 模块路径
export ZXING_WASM_PATH=/path/to/zxing.wasm

# 启用调试模式
export ZXING_DEBUG=true
```

### 工厂方法

```go
// 创建 CGO 版本
zx, err := zxing.NewCGO()

// 创建 WASM 版本
zx, err := zxing.NewWASM()

// 使用指定后端
zx, err := zxing.NewWithBackend(zxing.BackendWASM)
```

## 构建和部署

### 构建 WASM 版本

```bash
# 构建 Go WASM 程序
GOOS=js GOARCH=wasm go build -o app.wasm ./cmd/wasm-example/

# 复制 wasm_exec.js
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

### 构建 C++ WASM 模块

```bash
# 使用 Emscripten 编译 zxing
emcc -O3 -s WASM=1 \
     -s EXPORTED_FUNCTIONS='["_decode_image", "_encode_text"]' \
     -s MODULARIZE=1 \
     -s EXPORT_NAME='ZXingWASM' \
     zxing_sources.cpp -o zxing.js
```

### 在浏览器中使用

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>ZXing WASM Demo</title>
</head>
<body>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject)
            .then((result) => {
                go.run(result.instance);
            });
    </script>
</body>
</html>
```

## 性能优化

### 编译优化

1. **Go 编译优化**
   ```bash
   GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o app.wasm
   ```

2. **C++ 编译优化**
   ```bash
   emcc -O3 -flto --closure 1 -s WASM=1 ...
   ```

### 运行时优化

1. **预加载 WASM 模块**
2. **使用 Web Workers**
3. **内存池管理**
4. **批量处理**

## 错误处理

### 常见错误

1. **WASM 模块加载失败**
   ```go
   if err != nil {
       if strings.Contains(err.Error(), "WASM runtime not available") {
           // 回退到 CGO 实现
           config.Backend = zxing.BackendCGO
           zx, err = zxing.New(config)
       }
   }
   ```

2. **跨域问题**
   - 确保 WASM 文件可以被正确加载
   - 设置适当的 CORS 头

3. **内存不足**
   - 监控 WASM 内存使用
   - 及时释放资源

### 调试技巧

1. **启用调试模式**
   ```go
   config.Debug = true
   ```

2. **检查后端选择**
   ```go
   fmt.Printf("使用后端: %s\n", zx.GetBackend())
   ```

3. **性能分析**
   ```bash
   go test -bench=. -cpuprofile=cpu.prof
   ```

## 最佳实践

### 代码组织

1. **接口隔离**: 使用统一接口，隐藏实现细节
2. **配置管理**: 集中管理配置选项
3. **错误处理**: 提供详细的错误信息
4. **资源管理**: 及时释放资源

### 部署策略

1. **渐进式部署**: 先在测试环境验证
2. **回退机制**: 支持多种后端选择
3. **监控告警**: 监控性能和错误率
4. **版本管理**: 管理 WASM 模块版本

### 测试策略

1. **单元测试**: 测试各个组件
2. **集成测试**: 测试端到端流程
3. **性能测试**: 对比不同后端性能
4. **兼容性测试**: 测试不同环境

## 故障排除

### 编译问题

1. **构建标签错误**
   - 检查 `//go:build` 标签
   - 确保文件名正确

2. **依赖问题**
   - 运行 `go mod tidy`
   - 检查 Go 版本兼容性

### 运行时问题

1. **WASM 加载失败**
   - 检查文件路径
   - 验证 WASM 文件完整性

2. **性能问题**
   - 使用性能分析工具
   - 优化算法和数据结构

## 未来规划

1. **功能增强**
   - 支持更多条码格式
   - 添加批量处理功能
   - 优化内存使用

2. **性能优化**
   - 使用 SIMD 指令
   - 多线程支持
   - 缓存机制

3. **生态系统**
   - 提供更多示例
   - 集成到流行框架
   - 社区贡献指南