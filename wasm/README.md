# ZXing WASM 使用说明

## 概述

本项目提供了ZXingCPP的WebAssembly (WASM)支持，允许在浏览器环境中使用ZXing库进行条码解码和编码，而无需CGO环境。

## 文件结构

```
wasm/
├── README.md           # 本说明文档
├── test.html          # WASM测试页面
├── zxingwrapper.wasm  # 编译后的WASM模块
└── zxingwrapper.js    # WASM加载器脚本
```

## 构建WASM模块

### 前置要求

1. 安装Emscripten SDK
   ```bash
   # 克隆Emscripten SDK
   git clone https://github.com/emscripten-core/emsdk.git
   cd emsdk
   
   # 安装最新版本
   ./emsdk install latest
   ./emsdk activate latest
   
   # 设置环境变量
   source ./emsdk_env.sh
   ```

2. 确保已安装CMake和构建工具

### Windows平台构建

```powershell
# 使用PowerShell脚本
.\build_wasm.ps1

# 或手动构建
mkdir build-wasm
cd build-wasm
cmake -G "MinGW Makefiles" -DCMAKE_TOOLCHAIN_FILE="$env:EMSDK\upstream\emscripten\cmake\Modules\Platform\Emscripten.cmake" ..
cmake --build . --config Release
```

### Linux平台构建

```bash
# 使用Shell脚本
chmod +x build_wasm.sh
./build_wasm.sh

# 或手动构建
mkdir build-wasm
cd build-wasm
cmake -DCMAKE_TOOLCHAIN_FILE="$EMSDK/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake" ..
cmake --build . --config Release
```

## 使用方法

### 1. 在HTML页面中使用

```html
<!DOCTYPE html>
<html>
<head>
    <title>ZXing WASM Test</title>
</head>
<body>
    <script type="module">
        // 加载WASM模块
        async function loadWASM() {
            const response = await fetch('./zxingwrapper.wasm');
            const wasmBuffer = await response.arrayBuffer();
            
            const wasmModule = await WebAssembly.instantiate(wasmBuffer, {
                env: {
                    memory: new WebAssembly.Memory({ initial: 256 }),
                    abort: () => console.error('WASM abort called'),
                }
            });
            
            return wasmModule.instance.exports;
        }
        
        // 使用WASM模块
        async function main() {
            const zxing = await loadWASM();
            
            // 现在可以使用zxing模块的函数
            console.log('WASM module loaded:', zxing);
        }
        
        main();
    </script>
</body>
</html>
```

### 2. 在Go WASM项目中使用

```go
//go:build js && wasm

package main

import (
    "github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
    config := &zxing.Config{
        Backend:  zxing.BackendWASM,
        WASMPath: "./wasm/zxingwrapper.wasm",
        Debug:    true,
    }
    
    zx, err := zxing.New(config)
    if err != nil {
        panic(err)
    }
    defer zx.Close()
    
    // 使用ZXing功能
    // ...
}
```

## 支持的条码格式

- QR Code
- Code 128
- Code 39
- EAN-13
- EAN-8
- UPC-A
- UPC-E
- Data Matrix
- PDF417
- Aztec
- Codabar
- ITF
- MaxiCode

## API接口

### 解码函数

- `decode_image_data(data, width, height, channels)` - 解码图像数据
- `decode_image_file(filepath)` - 解码图像文件
- `decode_multiple_barcodes(data, width, height, channels)` - 解码多个条码

### 编码函数

- `encode_text_to_qr(text, width, height)` - 编码文本为QR码
- `encode_text_to_barcode(text, format, width, height)` - 编码文本为指定格式条码

### 工具函数

- `get_supported_formats()` - 获取支持的条码格式
- `get_last_error()` - 获取最后的错误信息

## 测试

1. 启动本地HTTP服务器（WASM需要通过HTTP协议加载）
   ```bash
   # Python 3
   python -m http.server 8000
   
   # Node.js
   npx http-server
   
   # Go
   go run -tags js,wasm test_wasm.go
   ```

2. 在浏览器中打开 `http://localhost:8000/wasm/test.html`

3. 测试各项功能：
   - 检查WASM模块状态
   - 生成QR码
   - 查看支持的格式
   - 上传并解码图像

## 性能特点

- **优势**：
  - 跨平台兼容性好
  - 无需CGO环境
  - 可在浏览器中运行
  - 支持现代Web标准

- **限制**：
  - 性能可能略低于CGO版本
  - 需要额外的WASM运行时
  - 模块大小较大

## 故障排除

### 常见问题

1. **WASM模块加载失败**
   - 检查文件路径是否正确
   - 确保通过HTTP协议访问（不是file://）
   - 检查浏览器控制台错误信息

2. **内存不足错误**
   - 增加WASM内存初始大小
   - 检查图像数据大小是否合理

3. **函数调用失败**
   - 确认WASM模块已正确加载
   - 检查函数名称是否正确
   - 验证参数类型和数量

### 调试技巧

1. 启用浏览器开发者工具
2. 查看控制台日志输出
3. 使用Network面板检查WASM文件加载
4. 在Sources面板中调试JavaScript代码

## 更新日志

- **v1.0.0**: 初始WASM支持
  - 基本的条码解码功能
  - QR码编码功能
  - 跨平台构建支持

## 许可证

本项目遵循与ZXingCPP相同的许可证。
