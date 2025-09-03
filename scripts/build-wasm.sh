#!/bin/bash

# ZXing WASM 构建脚本
# 用于构建 WebAssembly 版本的 ZXing

set -e

echo "开始构建 ZXing WASM 版本..."

# 检查必要的工具
check_tools() {
    echo "检查构建工具..."
    
    if ! command -v emcc &> /dev/null; then
        echo "错误: 未找到 Emscripten 编译器"
        echo "请先安装 Emscripten SDK:"
        echo "  git clone https://github.com/emscripten-core/emsdk.git"
        echo "  cd emsdk"
        echo "  ./emsdk install latest"
        echo "  ./emsdk activate latest"
        echo "  source ./emsdk_env.sh"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo "错误: 未找到 Go 编译器"
        exit 1
    fi
    
    echo "工具检查完成"
}

# 构建 WASM 模块
build_wasm() {
    echo "构建 WASM 模块..."
    
    cd wasm
    
    # 检查是否存在构建脚本
    if [ -f "build.sh" ]; then
        chmod +x build.sh
        ./build.sh
    else
        echo "警告: 未找到 wasm/build.sh，跳过 WASM 模块构建"
    fi
    
    cd ..
    
    echo "WASM 模块构建完成"
}

# 构建 Go 程序
build_go() {
    echo "构建 Go 程序..."
    
    # 构建示例程序
    echo "构建 WASM 示例程序..."
    cd cmd/wasm-example
    GOOS=js GOARCH=wasm go build -o ../../wasm/wasm-example.wasm .
    cd ../..
    
    # 构建其他程序（如果需要）
    echo "构建其他程序..."
    go build -o bin/zxing-cli ./cmd/zxing-cli/
    
    echo "Go 程序构建完成"
}

# 创建测试页面
create_test_page() {
    echo "创建测试页面..."
    
    cat > wasm/test.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>ZXing WASM 测试</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; }
        button { padding: 10px 20px; margin: 5px; }
        #output { background: #f5f5f5; padding: 10px; margin: 10px 0; }
        canvas { border: 1px solid #ccc; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>ZXing WebAssembly 测试</h1>
        
        <div class="section">
            <h2>编码测试</h2>
            <input type="text" id="textInput" placeholder="输入要编码的文本" value="Hello, ZXing WASM!">
            <button onclick="encodeText()">编码为二维码</button>
            <canvas id="qrCanvas" width="256" height="256"></canvas>
        </div>
        
        <div class="section">
            <h2>解码测试</h2>
            <input type="file" id="imageInput" accept="image/*">
            <button onclick="decodeImage()">解码图像</button>
        </div>
        
        <div class="section">
            <h2>输出</h2>
            <div id="output"></div>
        </div>
    </div>

    <script src="wasm_exec.js"></script>
    <script>
        let wasmModule;
        
        // 初始化 WASM
        async function initWasm() {
            const go = new Go();
            const result = await WebAssembly.instantiateStreaming(fetch("wasm-example.wasm"), go.importObject);
            go.run(result.instance);
            
            log("WASM 模块加载完成");
        }
        
        // 编码文本
        function encodeText() {
            const text = document.getElementById('textInput').value;
            if (!text) {
                log("请输入要编码的文本");
                return;
            }
            
            log(`编码文本: ${text}`);
            
            // 这里应该调用 WASM 函数
            // 由于这是示例，我们创建一个简单的测试图案
            const canvas = document.getElementById('qrCanvas');
            const ctx = canvas.getContext('2d');
            
            // 清空画布
            ctx.fillStyle = 'white';
            ctx.fillRect(0, 0, 256, 256);
            
            // 绘制简单的测试图案
            ctx.fillStyle = 'black';
            for (let i = 0; i < 16; i++) {
                for (let j = 0; j < 16; j++) {
                    if ((i + j) % 2 === 0) {
                        ctx.fillRect(i * 16, j * 16, 16, 16);
                    }
                }
            }
            
            log("编码完成（示例图案）");
        }
        
        // 解码图像
        function decodeImage() {
            const input = document.getElementById('imageInput');
            const file = input.files[0];
            
            if (!file) {
                log("请选择图像文件");
                return;
            }
            
            const reader = new FileReader();
            reader.onload = function(e) {
                const img = new Image();
                img.onload = function() {
                    log(`解码图像: ${file.name} (${img.width}x${img.height})`);
                    
                    // 这里应该调用 WASM 解码函数
                    log("解码结果: 示例文本 (QR_CODE)");
                };
                img.src = e.target.result;
            };
            reader.readAsDataURL(file);
        }
        
        // 日志输出
        function log(message) {
            const output = document.getElementById('output');
            const time = new Date().toLocaleTimeString();
            output.innerHTML += `<div>[${time}] ${message}</div>`;
            output.scrollTop = output.scrollHeight;
        }
        
        // 初始化
        initWasm().catch(err => {
            log(`WASM 初始化失败: ${err}`);
        });
    </script>
</body>
</html>
EOF

    # 复制 Go 的 wasm_exec.js
    if [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then
        cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm/
        echo "已复制 wasm_exec.js"
    else
        echo "警告: 未找到 wasm_exec.js"
    fi
    
    echo "测试页面创建完成: wasm/test.html"
}

# 主函数
main() {
    echo "ZXing WASM 构建脚本"
    echo "===================="
    
    check_tools
    build_wasm
    build_go
    create_test_page
    
    echo ""
    echo "构建完成！"
    echo ""
    echo "使用方法:"
    echo "1. 启动 HTTP 服务器: cd wasm && python -m http.server 8080"
    echo "2. 打开浏览器访问: http://localhost:8080/test.html"
    echo "3. 或者运行 Go 程序测试: ./bin/zxing-cli"
    echo ""
}

# 运行主函数
main "$@"