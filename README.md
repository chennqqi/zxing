# ZXing Go

这是一个基于 [zxing-cpp](https://github.com/zxing-cpp/zxing-cpp) 的 Go 语言条码识别库。

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

## 依赖要求

- Go 1.16 或更高版本
- CMake 3.10 或更高版本
- C++17 兼容的编译器
- zxing-cpp 库

## 安装依赖

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

## 构建

### Windows

```bash
# 运行构建脚本
.\build.bat
```

### Linux/macOS

```bash
# 添加执行权限
chmod +x build.sh

# 运行构建脚本
./build.sh
```

构建完成后，会在 `bin` 目录下生成以下文件：
- Windows: `zxing.dll`
- Linux: `libzxing.so`
- macOS: `libzxing.dylib`

## 使用示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/threatbook/zxing"
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

## 解码选项

```go
type DecodeOptions struct {
    Formats      BarcodeFormat // 要识别的条码格式
    TryHarder    bool          // 是否尝试更努力的解码
    TryRotate    bool          // 是否尝试旋转图像
    TryInvert    bool          // 是否尝试反转图像
    TryDownscale bool          // 是否尝试缩小图像
}
```

## 条码格式

```go
const (
    FormatNone      BarcodeFormat = iota
    FormatQRCode
    FormatAztec
    FormatCodabar
    FormatCode39
    FormatCode93
    FormatCode128
    FormatDataMatrix
    FormatEAN8
    FormatEAN13
    FormatITF
    FormatMaxiCode
    FormatPDF417
    FormatUPCA
    FormatUPCE
    FormatAll
)
```

## 许可证

MIT License