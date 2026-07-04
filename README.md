# ZXing Go

这是一个基于 [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp) 的 Go 语言条码识别库。

## 🚀 新特性：WASM 集成

现在支持 **WebAssembly (WASM)** 集成，提供无 CGO 依赖的替代方案！

- ✅ **无 CGO 依赖**: 通过 WASM 避免复杂的 C++ 编译
- ✅ **统一接口**: CGO 和 WASM 后端完全兼容
- ✅ **自动选择**: 根据环境自动选择最佳后端
- ✅ **跨平台**: 更好的跨平台兼容性

## 功能特性

- 支持多种条码格式：
  - QR Code
  - Aztec
  - Codabar
  - Code 39/93/128
  - Data Matrix
  - EAN-8/13
  - ITF
  - MaxiCode
  - PDF417
  - UPC-A/E
- 支持单条码和多条码识别
- 提供丰富的解码选项
- 支持错误处理
- **新增**: 支持 CGO 和 WASM 两种后端

## 依赖要求

### 传统 CGO 方式
- Go 1.18 或更高版本
- CMake 3.10 或更高版本
- C++17 兼容的编译器
- zxing-cpp 库

### 新的 WASM 方式
- Go 1.18 或更高版本
- 无需 CGO 和 C++ 编译器

## 快速开始

```bash
# 克隆项目（包含子模块）
git clone --recursive https://github.com/chennqqi/zxing.git
cd zxing

# 如果已经克隆过，初始化子模块
git submodule update --init --recursive

# 构建项目
go build ./pkg/zxing/
go build ./cmd/wasm-example/

# 运行示例
go run ./cmd/wasm-example/

# 运行测试
go test ./pkg/zxing/ -v
```

## 使用示例

### 传统 CGO 方式

```go
package main

import (
    "fmt"
    "log"

    "github.com/chennqqi/zxing"
)

func main() {
    // 创建默认选项
    options := zxing.NewDefaultOptions()
    if options == nil {
        log.Fatal("Failed to create default options")
    }

    // 设置只识别二维码
    options.Formats = zxing.FormatQRCode

    // 解码单个二维码
    result, err := zxing.Decode("test.png", options)
    if err != nil {
        log.Fatalf("Failed to decode: %v", err)
    }

    fmt.Printf("Decoded text: %s\n", result.Text)
    fmt.Printf("Format: %v\n", result.Format)
    fmt.Printf("Confidence: %.2f\n", result.Confidence)
}
```

### 新的统一接口（支持 WASM）

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    // 方法1: 自动选择后端
    zx, err := zxing.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer zx.Close()

    // 方法2: 明确使用 WASM 后端
    zx, err = zxing.NewWASM()
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

    // 解码图像
    result, err := zx.DecodeImage(context.Background(), img, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("解码结果: %s (格式: %s)\n", result.Text, result.Format)
}
```

## 配置选项

### 环境变量配置

```bash
# 指定使用 WASM 后端
export ZXING_BACKEND=wasm

# 指定 WASM 模块路径  
export ZXING_WASM_PATH=./wasm/zxing.wasm

# 启用调试模式
export ZXING_DEBUG=true

# 运行程序
go run ./cmd/wasm-example/
```

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

### 工厂方法

```go
// 自动选择后端
zx, err := zxing.New(nil)

// 使用 CGO 后端
zx, err := zxing.NewCGO()

// 使用 WASM 后端
zx, err := zxing.NewWASM()

// 使用指定后端
zx, err := zxing.NewWithBackend(zxing.BackendWASM)
```

## 构建选项

### 构建 WASM 版本

```bash
# 构建 Go WASM 程序
GOOS=js GOARCH=wasm go build -o wasm/app.wasm ./cmd/wasm-example/

# 复制 wasm_exec.js
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm/
```

### 构建 CGO 版本

```bash
# 构建 CGO 版本
CGO_ENABLED=1 go build -tags cgo ./cmd/wasm-example/
```

## 安装依赖（仅 CGO 方式需要）

### Windows

```bash
# 使用 vcpkg 安装 zxing-cpp
vcpkg install zxing-cpp:x64-windows
```

### Ubuntu/Debian

```bash
# 安装构建工具
sudo apt-get update
sudo apt-get install -y build-essential cmake

# 安装 zxing-cpp
sudo apt-get install -y libzxing-dev
```

### CentOS/RHEL

```bash
# 安装构建工具
sudo yum groupinstall "Development Tools"
sudo yum install cmake

# 安装 zxing-cpp
sudo yum install zxing-cpp-devel
```

### macOS

```bash
# 安装构建工具
brew install cmake

# 安装 zxing-cpp
brew install zxing-cpp
```

## 性能对比

| 后端 | 编译时间 | 运行性能 | 内存使用 | 跨平台性 | CGO 依赖 |
|------|----------|----------|----------|----------|----------|
| CGO  | 慢       | 最快     | 低       | 受限     | 是       |
| WASM | 中等     | 快       | 中等     | 优秀     | 否       |

## 项目结构

```
zxing/
├── cmd/
│   └── wasm-example/          # WASM 示例程序
├── pkg/
│   ├── zxing/                 # 统一接口层
│   └── wasm/                  # WASM 运行时
├── wasm/                      # WASM 构建文件
├── doc/                       # 文档
├── scripts/                   # 构建脚本
└── README.md
```

## 文档

- [WASM 集成指南](doc/wasm-integration-guide.md)
- [需求分析](doc/requirements-analysis.md)
- [开发需求](doc/requirements.md)

## 测试

```bash
# 运行单元测试
go test ./pkg/zxing/ -v

# 运行基准测试
go test -bench=. ./pkg/zxing/

# 测试覆盖率
go test -cover ./pkg/zxing/
```

## 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境设置

```bash
git clone --recursive https://github.com/chennqqi/zxing.git
cd zxing
git submodule update --init --recursive
go mod tidy
go test ./...
```

## 许可证

MIT License

## 相关链接

- [ZXing 官方项目](https://github.com/zxing/zxing)
- [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp)
- [WebAssembly](https://webassembly.org/)