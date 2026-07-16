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

## 2026-07-06: 修复 check-upstream workflow

- 修复版本比较逻辑：submodule 未检出时 `git describe --tags` 失败回退为 SHA，改为通过 GitHub API 解析 SHA 到 tag 名
- 修复 label 不存在导致 issue 创建失败：添加 `gh label create --force` 预创建 `upstream-update` label
- 修复 Node.js 20 弃用警告：所有 workflow 中 `actions/checkout@v4` 升级为 `@v5`

## 2026-07-06 (二): 修复 CI 构建失败

- 修复 `TestBackendSelection/CGO_backend` 测试失败：CGO_ENABLED=0 时应 skip 而非 Fatal
- 修复 Docker 构建中 `stb_image.h: No such file or directory`：CMakeLists.txt 中 zxingwrapper 目标缺少 stb 依赖
  - 根因：zxing-cpp 子目录通过 FetchContent 获取 stb，但 IMPORTED target 仅在子目录作用域可见
  - 修复：在顶层 CMakeLists.txt 中显式调用 `zxing_add_package_stb()` 并链接 `stb::stb` 到 zxingwrapper
  - 同步修复 `CMakeLists-wasm.txt`，并禁用不需要的 `ZXING_C_API`

## 相关文档

- `doc/windows-msvc-and-wasm-setup.md` - Windows平台MSVC和WASM构建环境说明
- `doc/build-progress.md` - 项目构建进度报告
- `doc/cli-build-and-test-summary.md` - CLI工具构建和测试总结
- `QRCODE_TEST_REPORT.md` - QR Code 识别测试报告

## 2026-07-04: 评审生产就绪设计规格

- 对 `docs/superpowers/specs/2026-07-04-production-ready-design.md` 进行设计评审
- 评审报告保存至 `docs/superpowers/reviews/2026-07-04-production-ready-design-review.md`
- 升级 zxing-cpp 从 v2.3.0 到 v3.0.2
- 增加 GitHub Action 检查上游新版本并通知

## 2026-07-05: 评审 `cmd/build/build_go.go` 实现

- 对新增的 `cmd/build/build_go.go` 及构建工具集成进行代码评审
- 评审报告保存至 `docs/superpowers/reviews/2026-07-05-cmd-build-implementation-review.md`

## 2026-07-05 (二): 复评 `cmd/build` 修复版

- 针对用户根据首轮评审修改后的 `cmd/build/build_go.go` 及 `env.go`/`test.go` 进行复评
- 评审报告保存至 `docs/superpowers/reviews/2026-07-05-cmd-build-implementation-review-2.md`

## 1.0.0 Review Fix - wazero runtime (2026-07-16)

- Fix WASM OOB: replace PNG+bump allocator path with raw-pixel decode_barcode_pixels ABI
- Add mutex for concurrent safety in Runtime and wasmZXing
- Propagate caller context.Context into wazero (WithCloseOnContextDone)
- Proper resource cleanup on init failure (close runtime/compiled module)
- Pass DecodeOptions (formats, try_harder, try_rotate, try_invert, try_downscale) through to WASM
- EncodeText returns explicit "not implemented" error
- Close returns underlying error, is idempotent
- Upgrade wazero v1.8.0 -> v1.12.0
- Add zxing_malloc/zxing_free C wrappers for reliable WASM export
- Rebuild wasm/zxingwrapper.wasm
