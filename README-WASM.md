# ZXing Go WASM 集成

本项目为 ZXing Go wrapper 添加了 WebAssembly (WASM) 支持，提供了无需 CGO 的替代方案。

## 特性

- 🚀 **无 CGO 依赖**: 通过 WASM 避免 CGO 编译复杂性
- 🔄 **统一接口**: 支持 CGO 和 WASM 两种后端，API 完全兼容
- 🎯 **自动选择**: 可根据环境自动选择最佳后端
- 📱 **跨平台**: WASM 版本支持更广泛的平台
- ⚡ **高性能**: 接近原生性能的 WASM 实现

## 快速开始

### 1. 安装依赖

```bash
# 安装 Emscripten (用于编译 WASM)
git clone https://github.com/emscripten-core/emsdk.git
cd emsdk
./emsdk install latest
./emsdk activate latest
source ./emsdk_env.sh
```

### 2. 构建项目

```bash
# 构建 WASM 版本
./scripts/build-wasm.sh

# 或者只构建 Go 部分
go build ./cmd/wasm-example/
```

### 3. 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    // 创建 WASM 版本的 ZXing 实例
    zx, err := zxing.NewWASM()
    if err != nil {
        log.Fatal(err)
    }
    defer zx.Close()

    // 编码文本为二维码
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

## 配置选项

### 环境变量

- `ZXING_BACKEND`: 指定后端类型 (`auto`, `cgo`, `wasm`)
- `ZXING_WASM_PATH`: WASM 模块路径
- `ZXING_DEBUG`: 启用调试模式 (`true`/`false`)

### 代码配置

```go
config := &zxing.Config{
    Backend:  zxing.BackendWASM,
    WASMPath: "path/to/zxing.wasm",
    Timeout:  30,
    Debug:    true,
}

zx, err := zxing.New(config)
```

## API 文档

### 主要接口

```go
type ZXing interface {
    // 解码图像
    DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error)
    
    // 解码字节数据
    DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error)
    
    // 编码文本为图像
    EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error)
    
    // 编码文本为字节数据
    EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error)
    
    // 关闭资源
    Close() error
    
    // 获取后端类型
    GetBackend() Backend
}
```

### 结果结构

```go
type Result struct {
    Text     string                 // 解码文本
    Format   string                 // 条码格式
    Points   []image.Point          // 位置点
    Metadata map[string]interface{} // 元数据
}
```

## 性能对比

| 后端 | 编译时间 | 运行性能 | 内存使用 | 跨平台性 |
|------|----------|----------|----------|----------|
| CGO  | 慢       | 最快     | 低       | 受限     |
| WASM | 中等     | 快       | 中等     | 优秀     |

## 构建选项

### 仅 WASM 版本

```bash
GOOS=js GOARCH=wasm go build -tags wasm ./cmd/wasm-example/
```

### 仅 CGO 版本

```bash
go build -tags cgo ./cmd/basic/
```

### 通用版本（自动选择）

```bash
go build ./cmd/wasm-example/
```

## 测试

### 单元测试

```bash
go test ./pkg/zxing/...
```

### 浏览器测试

```bash
cd wasm
python -m http.server 8080
# 访问 http://localhost:8080/test.html
```

### 性能测试

```bash
go test -bench=. ./pkg/zxing/
```

## 故障排除

### 常见问题

1. **WASM 模块加载失败**
   - 检查 WASM 文件路径是否正确
   - 确保在支持 WebAssembly 的环境中运行

2. **Emscripten 编译错误**
   - 确保已正确安装和配置 Emscripten
   - 检查 C++ 源码兼容性

3. **性能问题**
   - 尝试启用编译优化选项
   - 考虑使用 CGO 版本获得最佳性能

### 调试模式

```bash
export ZXING_DEBUG=true
export ZXING_BACKEND=wasm
go run ./cmd/wasm-example/
```

## 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境设置

```bash
git clone https://github.com/chennqqi/zxing.git
cd zxing
go mod tidy
./scripts/build-wasm.sh
```

## 许可证

本项目采用与原 ZXing 项目相同的许可证。

## 相关链接

- [ZXing 官方项目](https://github.com/zxing/zxing)
- [Emscripten 文档](https://emscripten.org/docs/)
- [WebAssembly 规范](https://webassembly.org/)