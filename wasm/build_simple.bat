@echo off
echo 开始构建 ZXing WASM 模块...

REM 检查 Emscripten 是否安装
where emcc >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到 Emscripten。请先安装 Emscripten SDK。
    echo 安装指南: https://emscripten.org/docs/getting_started/downloads.html
    exit /b 1
)

REM 创建输出目录
if not exist "output" mkdir output

REM 编译简单的 WASM 模块
echo 编译 C++ 源码为 WASM...
emcc wrapper.cpp -o output/zxing.js ^
    -s WASM=1 ^
    -s EXPORTED_FUNCTIONS="['_decode_image', '_encode_text', '_malloc', '_free']" ^
    -s EXPORTED_RUNTIME_METHODS="['ccall', 'cwrap']" ^
    -s ALLOW_MEMORY_GROWTH=1 ^
    -s MODULARIZE=1 ^
    -s EXPORT_NAME="ZXingModule" ^
    -O2

if %errorlevel% neq 0 (
    echo 编译失败！
    exit /b 1
)

echo WASM 模块构建成功！
echo 输出文件:
echo   - output/zxing.js
echo   - output/zxing.wasm

REM 复制到项目根目录
copy output\zxing.wasm ..\zxing.wasm >nul
copy output\zxing.js ..\zxing.js >nul

echo 文件已复制到项目根目录
echo 构建完成！