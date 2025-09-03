# ZXing Go WASM 集成

本项目为 ZXing Go wrapper 添加了 WebAssembly (WASM) 支持，提供了无需 CGO 的替代实现方案。

## 🎯 项目目标

- **消除 CGO 依赖**: 通过 WASM 方式集成 ZXing，避免 CGO 编译复杂性
- **跨平台兼容**: 支持更多平台和架构
- **统一接口**: 提供一致的 API，支持 CGO 和 WASM 两种后端
- **灵活配置**: 支持运行时后端选择和配置

## 📁 项目结构

```
├── pkg/
│   ├── zxing/              # 核心 ZXing 包
│   │   ├── interface.go    # 统一接口定义
│   │   ├── config.go       # 配置管理
│   │   ├── factory.go      # 工厂方法
│   │   └── universal_impl.go # 通用实现
│   └── wasm/               # WASM 运行时
│       ├── runtime_stub.go # 非 WASM 环境存根
│       └── runtime.go      # WASM 环境实现
├── wasm/                   # WASM 构建相关
│   ├── wrapper.cpp         # C++ 包装器
│   ├── build_simple.bat    # Windows 构建脚本
│   └── test_simple.html    # 测试页面
├── cmd/
│   └── wasm-example/       # 示例程序
└── doc/                    # 文档
```

## 🚀 快速开始

### 1. 构建和测试

```powershell
# 运行完整测试
.\test_build.ps1

# 或者手动执行
go build ./pkg/zxing/
go test ./pkg/zxing/ -v
go run ./cmd/wasm-example/
```

### 2. 基本使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    // 创建配置
    config := zxing.DefaultConfig()
    config.Backend = zxing.BackendWASM  // 或 BackendCGO, BackendAuto
    config.Debug = true
    
    // 创建 ZXing 实例
    zx, err := zxing.New(config)
    if err != nil {
        panic(err)
    }
    defer zx.Close()
    
    // 编码文本为二维码
    img, err := zx.EncodeText(context.Background(), "Hello, WASM!", nil)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("生成二维码: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
}
```

### 3. 环境变量配置

```bash
# 设置后端类型
export ZXING_BACKEND=wasm        # 或 cgo, auto

# 设置 WASM 文件路径
export ZXING_WASM_PATH=./wasm/zxing.wasm

# 启用调试模式
export ZXING_DEBUG=true

# 设置超时时间（秒）
export ZXING_TIMEOUT=30
```

## 🔧 WASM 模块构建

### 前置条件

1. 安装 [Emscripten SDK](https://emscripten.org/docs/getting_started/downloads.html)
2. 确保 `emcc` 命令可用

### 构建步骤

```bash
# Windows
cd wasm
.\build_simple.bat

# Linux/macOS
cd wasm
chmod +x build.sh
./build.sh
```

### 测试 WASM 模块

1. 构建完成后，打开 `wasm/test_simple.html`
2. 或启动本地服务器：
   ```bash
   python -m http.server 8000
   # 访问 http://localhost:8000/wasm/test_simple.html
   ```

## 📋 API 参考

### 核心接口

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
    
    // 获取后端类型
    GetBackend() Backend
    
    // 关闭资源
    Close() error
}
```

### 配置选项

```go
type Config struct {
    Backend  Backend  // 后端类型: cgo, wasm, auto
    WASMPath string   // WASM 文件路径
    Debug    bool     // 调试模式
    Timeout  int      // 超时时间（秒）
}

type DecodeOptions struct {
    TryHarder bool     // 尝试更努力解码
    Formats   []string // 支持的格式
}

type EncodeOptions struct {
    Width  int    // 图像宽度
    Height int    // 图像高度
    Format string // 编码格式
    Margin int    // 边距
}
```

## 🔄 后端选择策略

### 自动选择 (BackendAuto)

- **WASM 环境**: 自动使用 WASM 后端
- **其他环境**: 优先使用 CGO 后端

### 手动选择

```go
// 强制使用 WASM 后端
config.Backend = zxing.BackendWASM

// 强制使用 CGO 后端
config.Backend = zxing.BackendCGO
```

## 🧪 测试

```bash
# 运行所有测试
go test ./pkg/zxing/ -v

# 运行特定测试
go test ./pkg/zxing/ -run TestBackendSelection -v

# 运行基准测试
go test ./pkg/zxing/ -bench=. -v
```

## 📈 性能对比

| 后端类型 | 编译复杂度 | 运行性能 | 跨平台性 | 部署便利性 |
|---------|-----------|---------|---------|-----------|
| CGO     | 高        | 高      | 中      | 低        |
| WASM    | 中        | 中      | 高      | 高        |

## 🔍 故障排除

### 常见问题

1. **WASM 模块加载失败**
   - 检查 WASM 文件路径是否正确
   - 确保 WASM 文件存在且可读

2. **编译错误**
   - 确保 Go 版本 >= 1.19
   - 检查依赖包是否正确安装

3. **Emscripten 构建失败**
   - 确保 Emscripten SDK 正确安装
   - 检查 C++ 源码语法

### 调试模式

```go
config := zxing.DefaultConfig()
config.Debug = true  // 启用详细日志
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

## 📄 许可证

本项目采用与原 ZXing 项目相同的许可证。

## 🔗 相关链接

- [ZXing 官方项目](https://github.com/zxing/zxing)
- [Emscripten 文档](https://emscripten.org/docs/)
- [WebAssembly 规范](https://webassembly.org/)