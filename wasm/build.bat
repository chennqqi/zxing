@echo off
REM ZXing WASM 构建脚本 (Windows)
REM 使用 Emscripten 编译简化版 zxing 为 WebAssembly

echo 开始构建 ZXing WASM 模块...

REM 检查 Emscripten 环境
where emcc >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到 Emscripten 编译器
    echo 请先安装 Emscripten SDK:
    echo   git clone https://github.com/emscripten-core/emsdk.git
    echo   cd emsdk
    echo   emsdk install latest
    echo   emsdk activate latest
    echo   emsdk_env.bat
    exit /b 1
)

REM 创建构建目录
if not exist build mkdir build

echo 编译 C++ 源码为 WASM...

REM 编译简化版 zxing 为 WASM
emcc ^
    -O3 ^
    -s WASM=1 ^
    -s EXPORTED_RUNTIME_METHODS="[\"ccall\", \"cwrap\", \"getValue\", \"setValue\"]" ^
    -s EXPORTED_FUNCTIONS="[\"_decode_image_data\", \"_encode_text_to_qr\", \"_free_decode_result\", \"_free_encode_result\", \"_malloc\", \"_free\"]" ^
    -s ALLOW_MEMORY_GROWTH=1 ^
    -s MODULARIZE=1 ^
    -s EXPORT_NAME="ZXingWASM" ^
    -s ENVIRONMENT="web,worker" ^
    --bind ^
    -std=c++17 ^
    zxing_simple.cpp ^
    -o build/zxing.js

if %errorlevel% equ 0 (
    echo WASM 编译成功！
    
    REM 复制生成的文件
    copy build\zxing.js .
    copy build\zxing.wasm .
    
    echo 生成的文件:
    echo   - zxing.js ^(JavaScript 加载器^)
    echo   - zxing.wasm ^(WebAssembly 模块^)
    
    REM 显示文件大小
    for %%A in (zxing.wasm) do echo   - WASM 文件大小: %%~zA bytes
    
) else (
    echo WASM 编译失败！
    exit /b 1
)

echo 构建完成！
pause