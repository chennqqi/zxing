# ZXing Go

基于 [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp) 的 Go 语言条码识别库，支持 CGO 和 WASM (wazero) 双后端。

## 功能特性

- 支持多种条码格式：QR Code, Aztec, Codabar, Code 39/93/128, Data Matrix, EAN-8/13, ITF, MaxiCode, PDF417, UPC-A/E
- 支持单条码和多条码识别
- 双后端架构：
  - **CGO 后端**：原生 C++ 静态库，最高性能
  - **WASM 后端**：基于 [wazero](https://github.com/tetratelabs/wazero) 的纯 Go WASM 运行时，无 CGO 依赖
- 编译期后端选择：通过 Go build tags 自动选择
- 跨平台支持：Linux, Windows, macOS
- 预编译静态库：内置 Linux x64 和 Windows x64 的静态库，无需本地 C++ 编译环境

## 平台支持

| 平台 | CGO 后端 | WASM 后端 | 说明 |
|------|----------|----------|------|
| Linux x64 | ✅ | ✅ | 预编译静态库已包含 |
| Windows x64 | ✅ | ✅ | 预编译静态库已包含（需 MinGW-w64 GCC）|
| macOS | ❌ | ✅ | CGO 后端不支持，使用 WASM |
| js/wasm | ❌ | ✅ | 浏览器/Node.js 环境 |

> **Windows CGO 注意事项**: Windows 上的 CGO 后端需要 **MinGW-w64** GCC 工具链（非 MSVC）。
> 预编译静态库使用 MinGW-w64 编译，与 Go CGO 默认工具链兼容。
> 安装方法：`choco install mingw` 或从 [winlibs.com](https://winlibs.com/) 下载。

## 后端选择

后端通过编译期 build tags 自动选择：

| 条件 | 后端 | 说明 |
|------|------|------|
| `CGO_ENABLED=1` + Linux/Windows | CGO | 原生 C++ 静态库 |
| `CGO_ENABLED=0` 或 macOS | WASM (wazero) | 纯 Go WASM 运行时 |
| `GOOS=js GOARCH=wasm` | WASM (js) | 浏览器/Node.js 环境 |

也可通过 `Config.Backend` 手动指定：

```go
// 自动选择（默认）
zx, _ := zxing.New(nil)

// 强制使用 WASM
zx, _ := zxing.New(&zxing.Config{Backend: zxing.BackendWASM})

// 强制使用 CGO
zx, _ := zxing.New(&zxing.Config{Backend: zxing.BackendCGO})
```

## 快速开始

### WASM 后端（无需 CGO）

```bash
# 克隆项目
git clone --recursive https://github.com/chennqqi/zxing.git
cd zxing

# 使用 WASM 后端构建（默认）
CGO_ENABLED=0 go build ./pkg/zxing/

# 运行测试
CGO_ENABLED=0 go test ./pkg/zxing/ -v
```

### CGO 后端

```bash
# 使用构建工具（推荐）
go run ./cmd/build build-go

# 或手动构建
CGO_ENABLED=1 go build ./pkg/zxing/
```

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    // 自动选择后端
    zx, err := zxing.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer zx.Close()

    // 解码图像
    result, err := zx.DecodeImage(context.Background(), img, &zxing.DecodeOptions{
        TryHarder: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("解码结果: %s (格式: %s, 后端: %s)\n",
        result.Text, result.Format, zx.GetBackend())
}
```

## 构建工具

项目提供统一的构建工具 `cmd/build/`，替代所有碎片化脚本：

```bash
# 构建 Go 包
go run ./cmd/build build-go

# 构建 C++ 静态库
go run ./cmd/build build-lib

# 构建 WASM 模块（需要 Emscripten）
go run ./cmd/build build-wasm

# 构建全部
go run ./cmd/build build-all

# 同步 ZXing-CPP 头文件
go run ./cmd/build sync-headers

# 运行测试
go run ./cmd/build test

# 清理构建产物
go run ./cmd/build clean

# Docker 中构建 Linux 静态库（CentOS 7, glibc 2.17, GCC 10, C++20）
go run ./cmd/build docker-build
```

### 重新编译静态库

如果需要重新编译 C++ 静态库（例如更新了 zxing-cpp submodule）：

```bash
# Linux x64（使用 Docker, CentOS 7 + devtoolset-10, glibc 2.17 兼容）
docker build -t zxing-linux-build -f docker/Dockerfile.linux-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-linux-build \
  sh -c "/tmp/patch_using_enum.sh /workspace/zxing-cpp && \
  cd /tmp && rm -rf build && mkdir -p build && cd build && \
  cmake3 -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF \
  -DCMAKE_CXX_STANDARD=20 -DCMAKE_CXX_FLAGS=-fcoroutines /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/"

# Windows x64（使用 Docker 交叉编译, MinGW-w64）
docker build -t zxing-win-build -f docker/Dockerfile.win-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-win-build \
  sh -c "cd /tmp && mkdir -p build && cd build && \
  cmake -DCMAKE_TOOLCHAIN_FILE=/workspace/docker/mingw-w64-toolchain.cmake \
  -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/windows-x64/"
```

> **Linux 兼容性**: Linux 静态库在 CentOS 7 (glibc 2.17) 上编译，兼容 CentOS 7+、Ubuntu 16.04+、RHEL 7+、Debian 8+ 等所有使用 glibc 2.17+ 的发行版。
> GCC 10 需要 patch 脚本 (`docker/patch_using_enum.sh`) 替换 `using enum` 语法和 `-fcoroutines` 标志。

### 构建兼容旧系统的 CGO 二进制

如果需要 CGO 二进制也兼容旧系统（glibc 2.17+），**必须在 CentOS 7 容器内完成全部构建**（包括 `go build`）：

```bash
# 在 CentOS 7 容器内完成 C++ 静态库 + Go CGO 二进制构建
docker build -t zxing-linux-build -f docker/Dockerfile.linux-build docker/
docker run --rm -v "$PWD":/workspace:Z zxing-linux-build sh -c "
  /tmp/patch_using_enum.sh /workspace/zxing-cpp && \
  cd /tmp && rm -rf build && mkdir -p build && cd build && \
  cmake3 -DCMAKE_BUILD_TYPE=Release -DBUILD_STATIC_LIB=ON -DBUILD_SHARED_LIBS=OFF \
  -DCMAKE_CXX_STANDARD=20 -DCMAKE_CXX_FLAGS=-fcoroutines /workspace && \
  make -j\$(nproc) && cp lib/libZXing.a lib/libzxingwrapper.a /workspace/lib/linux-x64/ && \
  cd /workspace && \
  CGO_ENABLED=1 \
  CGO_CFLAGS='-I/workspace/include' \
  CGO_CXXFLAGS='-std=c++20 -I/workspace/include' \
  CGO_LDFLAGS='-L/workspace/lib/linux-x64 -lzxingwrapper -lZXing -lstdc++ -lm' \
  go build -buildvcs=false -o zxing-cli ./cmd/zxing-cli
"
```

> **为什么 `go build` 也要在容器内？** 静态库 (.a) 本身没有 glibc 版本标记，但 `go build` 链接最终二进制时，链接器会从宿主系统的 glibc 解析符号版本。在 glibc 2.39 的宿主上链接，即使静态库是 CentOS 7 编译的，最终二进制仍会要求 GLIBC_2.38。在 CentOS 7 内完成全部构建，最高只要求 GLIBC_2.14。

## 环境变量配置

```bash
# 指定后端
export ZXING_BACKEND=auto    # auto | cgo | wasm

# 指定 WASM 模块路径
export ZXING_WASM_PATH=wasm/zxingwrapper.wasm

# 启用调试模式
export ZXING_DEBUG=true

# 超时时间（秒）
export ZXING_TIMEOUT=30
```

## API

### ZXing 接口

```go
type ZXing interface {
    DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error)
    DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error)
    EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error)
    EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error)
    Close() error
    GetBackend() Backend
}
```

### 工厂方法

```go
// 自动选择后端
zx, err := zxing.New(nil)

// 指定 CGO 后端
zx, err := zxing.NewCGO(config)

// 指定 WASM 后端
zx, err := zxing.NewWASM(config)
```

## 项目结构

```
zxing/
├── cmd/
│   ├── build/                 # 统一构建工具
│   ├── zxing-cli/             # 命令行工具
│   └── server/                # HTTP 服务
├── pkg/
│   ├── zxing/                 # 统一接口层
│   │   ├── interface.go       # ZXing 接口定义
│   │   ├── config.go          # 配置
│   │   ├── factory.go         # 后端工厂
│   │   ├── cgo_binding_linux.go   # Linux CGO 绑定
│   │   ├── cgo_binding_windows.go # Windows CGO 绑定
│   │   ├── cgo_impl.go        # CGO 实现
│   │   ├── cgo_stub.go        # CGO stub（非 CGO 平台）
│   │   ├── wasm_impl.go       # wazero WASM 实现
│   │   ├── wasm_impl_js.go    # js/wasm 实现
│   │   └── wasm_stub.go       # WASM stub（CGO 平台）
│   └── wasm/                  # WASM 运行时
│       ├── runtime_wazero.go  # wazero 运行时
│       ├── runtime_js.go      # js/wasm 运行时
│       └── runtime_stub.go    # 运行时 stub
├── include/                   # C/C++ 头文件
├── lib/                       # 预编译静态库
├── wasm/                      # WASM 模块
├── docker/                    # Docker 构建环境
├── src/                       # C++ wrapper 源码
└── CMakeLists.txt             # CMake 构建配置
```

## 性能对比

| 后端 | 编译时间 | 运行性能 | 内存使用 | 跨平台性 | CGO 依赖 |
|------|----------|----------|----------|----------|----------|
| CGO  | 慢       | 最快     | 低       | 受限     | 是       |
| WASM | 快       | 快       | 中等     | 优秀     | 否       |

## 测试

```bash
# WASM 后端测试
CGO_ENABLED=0 go test ./pkg/zxing/ -v

# CGO 后端测试
CGO_ENABLED=1 go test ./pkg/zxing/ -v

# 构建工具测试
go test ./cmd/build/ -v

# WASM 运行时测试
CGO_ENABLED=0 go test ./pkg/wasm/ -v
```

## 许可证

MIT License

## 相关链接

- [ZXing 官方项目](https://github.com/zxing/zxing)
- [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp)
- [wazero](https://github.com/tetratelabs/wazero)
- [WebAssembly](https://webassembly.org/)
