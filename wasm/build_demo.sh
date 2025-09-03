#!/bin/bash

# 演示版构建脚本 - 创建模拟的 WASM 文件用于测试

echo "创建演示版 WASM 文件..."

# 创建一个简单的 JavaScript 模拟器
cat > zxing.js << 'EOF'
// ZXing WASM 模拟器 - 用于演示和测试
function ZXingWASM() {
    return new Promise((resolve) => {
        // 模拟异步加载
        setTimeout(() => {
            const module = {
                // 模拟编码函数
                encode_text_to_qr: function(text, width, height) {
                    console.log(`模拟编码: "${text}" ${width}x${height}`);
                    
                    const data = new Array(width * height);
                    
                    // 生成简单的测试图案
                    for (let y = 0; y < height; y++) {
                        for (let x = 0; x < width; x++) {
                            const idx = y * width + x;
                            
                            // 创建边框和简单图案
                            if (x < 10 || x >= width - 10 || y < 10 || y >= height - 10) {
                                data[idx] = 0; // 黑色边框
                            } else if ((x + y) % 20 < 10) {
                                data[idx] = 0; // 黑色
                            } else {
                                data[idx] = 255; // 白色
                            }
                        }
                    }
                    
                    return {
                        success: true,
                        width: width,
                        height: height,
                        data: {
                            size: () => data.length,
                            get: (i) => data[i]
                        },
                        error_code: 0,
                        error_message: ""
                    };
                },
                
                // 模拟解码函数
                decode_image_data: function(dataPtr, width, height, channels) {
                    console.log(`模拟解码: ${width}x${height}, channels=${channels}`);
                    
                    // 简单的模式检测
                    const hasPattern = Math.random() > 0.3; // 70% 成功率
                    
                    if (hasPattern) {
                        return {
                            success: true,
                            text: "Demo: Hello from WASM!",
                            format: "QR_CODE",
                            error_code: 0,
                            error_message: ""
                        };
                    } else {
                        return {
                            success: false,
                            text: "",
                            format: "",
                            error_code: 1,
                            error_message: "No barcode pattern detected"
                        };
                    }
                },
                
                // 模拟内存管理
                _malloc: function(size) {
                    return new ArrayBuffer(size);
                },
                
                _free: function(ptr) {
                    // 模拟释放内存
                },
                
                HEAPU8: {
                    set: function(data, offset) {
                        // 模拟内存设置
                    }
                }
            };
            
            resolve(module);
        }, 500);
    });
}

// 导出模块
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ZXingWASM;
} else if (typeof window !== 'undefined') {
    window.ZXingWASM = ZXingWASM;
}
EOF

# 创建一个空的 WASM 文件（用于演示）
echo -n "" > zxing.wasm

echo "演示文件创建完成:"
echo "  - zxing.js (JavaScript 模拟器)"
echo "  - zxing.wasm (空文件，仅用于演示)"
echo ""
echo "可以打开 test.html 进行测试"