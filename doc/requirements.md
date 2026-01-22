# ZXing Go Wrapper 项目需求文档

## 项目概述
本项目旨在实现Go语言包装的ZXingCPP库，提供二维码扫描功能。

## 核心需求
1. 基于ZXingCPP实现Go版本的二维码扫描
2. ZXingCPP使用上游https://github.com/zxing-cpp/zxing-cpp 稳定版本的代码，当前为v2.3.0
3. CGO方式编译，将zxing-cpp编译为静态库存放于lib目录进行链接
4. WASM方式编译，将zxing-cpp编译WASM，以实现无CGO依赖调用zxing-cpp
5. 同时支持Windows/Linux平台
6. 考虑到后续编译的便利性，保存编译好的Windows/Linux lib文件、WASM文件到项目中

## 技术栈
- Go语言
- ZXingCPP C++库 (v2.3.0)
- CGO
- WebAssembly (WASM)
- CMake构建系统

## 功能特性
- 支持多种条码格式（QR Code, Code128, EAN等）
- 图像文件解码
- 批量解码支持
- 可配置的解码选项

## 项目阶段
### 阶段1：代码修改 ✅
1. 阅读当前项目中的代码，对无用代码、文档进行清理、删除
2. 分析当前代码是否可以完全实现目标，如不能，则补充代码
3. 补充WASM支持，实现完整的WebAssembly后端

### 阶段2：Windows平台编译
1. 检查Windows平台编译所需依赖
2. 安装Windows平台编译依赖
3. Windows平台编译
4. 验证编译结果；如果有问题进行修改直到完成
5. 进行单元测试，使用实际的数据进行二维码扫描识别
6. 保存编译结果静态库、WASM文件等

### 阶段3：Linux平台编译
1. 检查Linux平台编译所需依赖
2. 安装Linux平台编译依赖
3. Linux平台编译
4. 验证编译结果；如果有问题进行修改直到完成
5. 进行单元测试，使用实际的数据进行二维码扫描识别
6. 保存编译结果结果静态库、WASM文件等

### 阶段4：回归测试
回归测试Linux平台修改后是否破坏Windows平台，如破坏则需要回退到阶段3

### 阶段5：性能测试
1. 编写性能测试用例
2. 使用实际的数据分别测试两种方式的性能差异

### 阶段6：总结
1. 整理项目文档、代码、脚本，移除无关文件
2. 总结使用说明，更新README.md

## 历史需求记录
### 2024年需求
- 为zxing的Go wrapper项目添加WASM方式集成，通过WASM方式Go集成zxing，避免使用CGO
- 当前项目通过两种方式wrap zxingcpp：CGO方式和WASM方式
- 已完成编译测试和真实数据测试

### 2024年12月需求分析
- 分析项目目标达成度，发现以下未完成任务：
  1. CMakeLists.txt生成的是动态库，需要改为静态库
  2. 缺少lib目录来存放编译好的静态库文件
  3. 缺少编译好的Windows/Linux静态库和WASM文件
  4. CGO实现中有TODO标记，需要完善
  5. 需要创建过程性脚本来编译和保存静态库
- 详细分析见：doc/project-goal-analysis.md