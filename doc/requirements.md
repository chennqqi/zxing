# 项目需求文档

## 项目目标

实现Go语言包装的ZXingCPP库，实现二维码扫描功能。

## 核心目标

1. 基于ZXingCPP实现Go版本的二维码扫描
2. ZXingCPP使用上游https://github.com/zxing-cpp/zxing-cpp 稳定版本的代码，当前为v2.3.0
3. cgo方式编译，将zxing-cpp编译为静态库存放于lib目录进行链接
4. wasm方式编译，将zxing-cpp编译wasm，以实现无cgo依赖调用zxing-cpp
5. 同时支持windows/linux平台
6. 考虑到后续编译的便利性，保存编译好的windows/linux lib文件、wasm文件到项目中
7. 构建脚本不应依赖sudo权限，禁止安装库文件到系统目录，应使用本地构建和链接

## 项目阶段

### 阶段1：代码修改 ✅
1. 阅读当前项目中的代码，对无用代码、文档进行清理、删除 ✅
2. 分析当前代码是否可以完全实现目标，如不能，则补充代码 ✅

### 阶段2：windows平台编译 ⏳
1. 检查windows平台编译所需依赖 ✅
2. 安装windows平台编译依赖 ⏳
   - MSVC: ✅ 已安装（用于编译静态库）
   - MinGW-w64: ⏳ 待安装（用于CGO链接）
   - Emscripten SDK: ✅ 已安装（用于WASM编译）
3. windows平台编译 ⏳
   - 静态库: ✅ 已完成
   - CGO编译: ⏳ 需要gcc
   - WASM编译: ⏳ 准备开始
4. 验证编译结果 ⏳
5. 进行单元测试，使用实际的数据进行二维码扫描识别 ⏳
6. 保存编译结果静态库、wasm文件等 ⏳
   - 静态库: ✅ 已保存
   - WASM: ⏳ 待编译

### 阶段3：linux平台编译 ⏳
1. 检查linux平台编译所需依赖 ⏳
2. 安装linux平台编译依赖 ⏳
3. linux平台编译 ⏳
4. 验证编译结果 ⏳
5. 进行单元测试，使用实际的数据进行二维码扫描识别 ⏳
6. 保存编译结果结果静态库、wasm文件等 ⏳

### 阶段4: 回归测试 ⏳
回归测试linux平台修改后是否破坏windowws平台，如破坏则需要回退到阶段3

### 阶段5：性能测试 ⏳
1. 编写性能测试用例 ⏳
2. 使用实际的数据分别测试两种方式的性能差异 ⏳

### 阶段6：总结 ⏳
1. 整理项目文档、代码、脚本，移除无关文件 ⏳
2. 总结使用说明，更新README.md ⏳

## 当前问题

### 1. Windows平台CGO编译

**问题**: CGO在Windows上需要gcc编译器，但当前环境只有MSVC

**解决方案**:
- 安装MinGW-w64（推荐）
- 或配置CGO使用MSVC（不推荐，配置复杂）

**详细说明**: 见 `doc/windows-msvc-and-wasm-setup.md`

### 2. WASM构建环境

**问题**: 需要安装Emscripten SDK才能编译WASM模块

**解决方案**: 
- ✅ 已安装Emscripten SDK (v4.0.23)
- 下一步：执行WASM构建脚本

**详细说明**: 见 `doc/windows-msvc-and-wasm-setup.md`

## 测试任务

### 2026-01-25: QR Code 识别效果验证测试
- **任务**: 使用编译好的程序测试 data/images 中的图片，验证程序 QRCode 识别效果
- **测试结果**: 
  - 总图片数: 1,745 张
  - 成功识别: 1,593 张 (91.2%)
  - 失败: 152 张 (8.8%，均为图片文件损坏或格式不支持)
  - 识别准确率: 对于有效图片达到 100%
- **详细报告**: 见 `QRCODE_TEST_REPORT.md`

## 2026-07-04: 将 zxing-cpp 改为 git submodule
- 移除项目中内嵌的 zxing-cpp 源码副本，改为 git submodule 引用 https://github.com/zxing-cpp/zxing-cpp.git (v2.3.0)
- 更新构建脚本，使用 `git submodule update --init --recursive` 替代手动 git clone

## 相关文档

- `doc/windows-msvc-and-wasm-setup.md` - Windows平台MSVC和WASM构建环境说明
- `doc/build-progress.md` - 项目构建进度报告
- `doc/cli-build-and-test-summary.md` - CLI工具构建和测试总结
- `QRCODE_TEST_REPORT.md` - QR Code 识别测试报告
