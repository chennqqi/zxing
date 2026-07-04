# Production-Ready zxing Go Library 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 zxing Go wrapper 重构为生产可用，支持 Windows/Linux 双平台、CGO link 和 WASM no-CGO 双后端，提供统一构建工具和预编译库。

**架构：** 预编译静态库提交到仓库（Linux Docker CentOS 7 构建，Windows MinGW-w64 构建），clone 后直接 `go build` 可用。`CGO_ENABLED=0` 时通过 wazero 纯 Go 运行时加载 WASM 文件。Go 构建工具 `cmd/build/` 替代所有碎片化脚本。编译期 build tags 决定后端，无运行时切换。

**技术栈：** Go 1.24, CGO, wazero v1.8.0, CMake, Docker (CentOS 7), MinGW-w64, EMSDK

---

## 文件结构

### 新建文件

| 文件 | 职责 |
|------|------|
| `pkg/zxing/cgo_binding_linux.go` | `//go:build cgo && linux` — Linux CGO C 绑定 |
| `pkg/zxing/cgo_binding_windows.go` | `//go:build cgo && windows` — Windows CGO C 绑定 |
| `pkg/zxing/wasm_impl.go` | `//go:build !cgo \|\| !(linux \|\| windows)` — wazero WASM 后端 |
| `pkg/zxing/wasm_stub.go` | `//go:build cgo && (linux \|\| windows) && !(js && wasm)` — WASM stub |
| `pkg/wasm/runtime_wazero.go` | `//go:build !cgo \|\| !(linux \|\| windows)` — wazero 运行时 |
| `pkg/wasm/runtime_js.go` | `//go:build js && wasm` — 浏览器运行时（从 runtime.go 重命名） |
| `cmd/build/main.go` | 构建工具入口 |
| `cmd/build/build_lib.go` | build-lib 子命令 |
| `cmd/build/build_wasm.go` | build-wasm 子命令 |
| `cmd/build/build_go.go` | build-go 子命令 |
| `cmd/build/test.go` | test 子命令 |
| `cmd/build/clean.go` | clean 子命令 |
| `cmd/build/docker_build.go` | docker-build 子命令 |
| `cmd/build/sync_headers.go` | sync-headers 子命令 |
| `cmd/build/env.go` | CGO 环境变量工具函数 |
| `cmd/build/env_test.go` | env 工具函数单元测试 |
| `docker/Dockerfile.linux-build` | CentOS 7 构建环境 |
| `lib/BUILDINFO.md` | 预编译库审计信息 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `pkg/zxing/cgo_impl.go` | build tag → `cgo && (linux \|\| windows)`，移除 C import |
| `pkg/zxing/cgo_stub.go` | build tag → `!cgo \|\| !(linux \|\| windows)` |
| `pkg/zxing/wasm_impl_js.go` | 从 `wasm_impl.go` 重命名 |
| `pkg/zxing/factory.go` | 简化 newAuto |
| `pkg/zxing/universal_impl.go` | 删除运行时切换，按 build tag 分拆 |
| `pkg/zxing/zxing_test.go` | 更新测试 |
| `pkg/wasm/runtime_stub.go` | 更新 build tag |
| `go.mod` | 添加 wazero |
| `README.md` | 重写 |
| `.gitignore` | 更新 |
| `CMakeLists-wasm.txt` | 更新导出函数 |
| `src/zxing.cpp` | 添加 `decode_barcode_data` 函数 |
| `include/zxing.h` | 添加 `decode_barcode_data` 声明 |

### 删除文件

`zxing.go`, `build.sh`, `build.bat`, `build.ps1`, `build_all.ps1`, `build_wasm.ps1`, `build_wasm.sh`, `build_wasm_demo.ps1`, `scripts/build_static_linux.sh`, `scripts/build_static_windows.ps1`, `scripts/build_wasm_save.sh`, `scripts/build_wasm_save.ps1`, `test_build.ps1`, `test_integration.ps1`, `test_cmake.sh`, `Makefile`, `lib64/`

---

## 任务 0：PoC 验证 — wazero 加载 WASM

**目标：** 验证 wazero 能加载 `wasm/zxingwrapper.wasm`。

**文件：** 创建 `pkg/wasm/runtime_wazero.go`, `pkg/wasm/runtime_wazero_test.go`

- [ ] **步骤 1：添加 wazero 依赖**

```bash
go get github.com/tetratelabs/wazero@v1.8.0
```

- [ ] **步骤 2：编写 PoC 测试**

创建 `pkg/wasm/runtime_wazero_test.go`（`//go:build !cgo || !(linux || windows)`），测试 `NewRuntime()` → `Initialize(ctx, "../../wasm/zxingwrapper.wasm")` → `IsReady()` → `Close()`

- [ ] **步骤 3：编写 PoC runtime**

创建 `pkg/wasm/runtime_wazero.go`（`//go:build !cgo || !(linux || windows)`），包含 `Runtime` struct（`wazero.Runtime`, `api.Module`, `bool`），`DecodeResult`/`EncodeResult` 类型，`NewRuntime()`/`Initialize()`/`IsReady()`/`Close()` 方法。`DecodeImage`/`EncodeText` 暂返回 stub。

- [ ] **步骤 4：运行测试**

```bash
CGO_ENABLED=0 go test ./pkg/wasm/ -v -run TestWazeroLoadAndDecode
```
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add pkg/wasm/runtime_wazero.go pkg/wasm/runtime_wazero_test.go go.mod go.sum
git commit -m "poc: verify wazero can load zxingwrapper.wasm"
```

---

## 任务 1：重构 CGO 绑定 — 按平台分离

**目标：** CGO `#cgo` 指令按平台分离，指向各自预编译库路径。

**文件：** 创建 `cgo_binding_linux.go`, `cgo_binding_windows.go`；重写 `cgo_impl.go`, `cgo_stub.go`；删除 `cgo_binding.go`

- [ ] **步骤 1：创建 `cgo_binding_linux.go`**

`//go:build cgo && linux`，包含 `#cgo CFLAGS: -I${SRCDIR}/../../include`，`#cgo LDFLAGS: -L${SRCDIR}/../../lib/linux-x64 -lzxingwrapper -lZXing -lstdc++ -lm`，`#include "zxing.h"`，`import "C"`，以及 `BarcodeFormat` 类型、所有格式常量、`CGODecodeOptions`/`CGODecodeResult` 类型、`NewDefaultOptions()`/`Decode()`/`DecodeMulti()`/`boolToInt()` 函数签名。`Decode`/`DecodeMulti` 调用 `decodeCGO`/`decodeMultiCGO`（在 `cgo_impl.go` 中实现）。

- [ ] **步骤 2：创建 `cgo_binding_windows.go`**

同上但 `//go:build cgo && windows`，`#cgo LDFLAGS: -L${SRCDIR}/../../lib/windows-x64 -lzxingwrapper -lZXing -lstdc++`（无 `-lm`）。

- [ ] **步骤 3：重写 `cgo_impl.go`**

`//go:build cgo && (linux || windows)`，移除 `import "C"` 和类型常量（已移到 binding 文件）。保留 `barcodeFormatString()`、`decodeCGO()`、`decodeMultiCGO()`、`cgoZXing` struct 及其方法、`decodeWithCGOImpl()`/`encodeWithCGOImpl()`。这些函数使用 `C.` 引用，因为 binding 文件已 `import "C"`。

- [ ] **步骤 4：重写 `cgo_stub.go`**

`//go:build !cgo || !(linux || windows)`，包含 `BarcodeFormat` 类型及常量（stub 值）、`String()` 方法、`CGODecodeOptions`/`CGODecodeResult` stub 类型、`NewDefaultOptions()` 返回 nil、`Decode()`/`DecodeMulti()` 返回错误、`decodeWithCGOImpl()`/`encodeWithCGOImpl()` 返回错误、`boolToInt()`。

- [ ] **步骤 5：删除 `cgo_binding.go`**

```bash
git rm pkg/zxing/cgo_binding.go
```

- [ ] **步骤 6：验证 CGO 构建**

```bash
CGO_ENABLED=1 CGO_CFLAGS="-I$(pwd)/include" CGO_CXXFLAGS="-std=c++17 -I$(pwd)/include" \
CGO_LDFLAGS="-L$(pwd)/lib -lzxingwrapper -lZXing -lstdc++ -lm" go build ./pkg/zxing/
```

- [ ] **步骤 7：验证非 CGO 构建**

```bash
CGO_ENABLED=0 go build ./pkg/zxing/
```

- [ ] **步骤 8：Commit**

```bash
git add pkg/zxing/cgo_binding_linux.go pkg/zxing/cgo_binding_windows.go pkg/zxing/cgo_impl.go pkg/zxing/cgo_stub.go
git commit -m "refactor: split CGO bindings by platform (linux/windows)"
```

---

## 任务 2：添加 `decode_barcode_data` 到 WASM wrapper

**目标：** wazero 无 MEMFS，需要 WASM 导出接受原始图片字节数据的函数。

**文件：** 修改 `include/zxing.h`, `src/zxing.cpp`, `CMakeLists-wasm.txt`

- [ ] **步骤 1：在 `include/zxing.h` 添加声明**

在 `get_last_error()` 声明前添加：
```c
DecodeResult* decode_barcode_data(const unsigned char* file_data, int file_size, const DecodeOptions* options);
```

- [ ] **步骤 2：在 `src/zxing.cpp` 添加实现**

在文件末尾添加 `decode_barcode_data` 函数，使用 `stbi_load_from_memory` 加载图片数据，创建 `ImageView`，调用 `ReadBarcodes`，返回 `DecodeResult*`。函数签名：`DecodeResult* decode_barcode_data(const uint8_t* file_data, int file_size, const DecodeOptions* options)`。逻辑：stbi_load_from_memory → ImageView → ReaderOptions → ReadBarcodes → 分配并填充 DecodeResult → 返回。

- [ ] **步骤 3：更新 `CMakeLists-wasm.txt` 导出函数**

在 `EXPORTED_FUNCTIONS` 中添加 `"_decode_barcode_data"`。在 `EXPORTED_RUNTIME_METHODS` 中确保包含 `getValue`、`setValue`、`UTF8ToString`。

- [ ] **步骤 4：重新构建 WASM**

```bash
rm -rf build-wasm && mkdir build-wasm && cd build-wasm
cp ../CMakeLists-wasm.txt ../CMakeLists.txt
emcmake cmake .. -G "Unix Makefiles" -DCMAKE_BUILD_TYPE=Release
emmake make -j$(nproc)
cp bin/zxingwrapper.js bin/zxingwrapper.wasm ../wasm/
cd .. && git checkout CMakeLists.txt  # restore original
```

- [ ] **步骤 5：Commit**

```bash
git add include/zxing.h src/zxing.cpp CMakeLists-wasm.txt wasm/zxingwrapper.wasm wasm/zxingwrapper.js
git commit -m "feat: add decode_barcode_data for wazero raw data decode"
```

---

## 任务 3：完善 wazero WASM 后端

**目标：** 完善 `runtime_wazero.go` 的 `DecodeImage`，通过 `decode_barcode_data` 实现真实解码。

**文件：** 修改 `pkg/wasm/runtime_wazero.go`；测试 `pkg/wasm/runtime_wazero_test.go`

- [ ] **步骤 1：完善 `DecodeImage` 方法**

在 `runtime_wazero.go` 中实现：将 RGBA data 用 Go `image/png` 编码为 PNG → 在 WASM 中 `malloc` 内存写入 PNG 数据 → 调用 `create_default_options` → 调用 `decode_barcode_data(pngPtr, pngLen, optsPtr)` → 从 WASM 内存读取 `DecodeResult` struct（text 指针 + format + confidence）→ 读取 text 字符串 → `free_result` + `free_options` + `free`。需要辅助函数 `cgoString([]byte) string` 读取 null-terminated string，`encodeRGBAtoPNG([]byte, int, int) ([]byte, error)` 编码 PNG。

- [ ] **步骤 2：编写真实图片解码测试**

在 `runtime_wazero_test.go` 中添加 `TestWazeroDecodeQRCode`：加载 `../../data/qrcode_www.bing.com.png`，读取为 `image.Image`，转 RGBA bytes，调用 `DecodeImage`，验证 `result.Success == true` 且 `result.Text` 非空。

- [ ] **步骤 3：运行测试**

```bash
CGO_ENABLED=0 go test ./pkg/wasm/ -v -run TestWazeroDecodeQRCode
```

- [ ] **步骤 4：Commit**

```bash
git add pkg/wasm/runtime_wazero.go pkg/wasm/runtime_wazero_test.go
git commit -m "feat: implement wazero decode via decode_barcode_data"
```

---

## 任务 4：重构 `pkg/zxing/` WASM 后端和 factory

**目标：** 将 WASM 后端从 `//go:build js && wasm` 改为 `//go:build !cgo || !(linux || windows)`，简化 factory。

**文件：** 重命名 `wasm_impl.go` → `wasm_impl_js.go`；创建 `wasm_impl.go`（wazero）；创建 `wasm_stub.go`；修改 `factory.go`；修改 `universal_impl.go`；重命名 `pkg/wasm/runtime.go` → `runtime_js.go`；更新 `pkg/wasm/runtime_stub.go`

- [ ] **步骤 1：重命名现有 WASM 文件**

```bash
git mv pkg/zxing/wasm_impl.go pkg/zxing/wasm_impl_js.go
git mv pkg/wasm/runtime.go pkg/wasm/runtime_js.go
```

- [ ] **步骤 2：创建 `pkg/zxing/wasm_impl.go`**

`//go:build !cgo || !(linux || windows)`，实现 `wasmZXing` struct（config + `*wasm.Runtime`），`DecodeImage`/`DecodeBytes`/`EncodeText`/`EncodeToBytes`/`Close`/`GetBackend` 方法。`DecodeBytes` 将 RGBA 数据传给 `runtime.DecodeImage`。`EncodeText` 调用 `runtime.EncodeText`。`GetBackend` 返回 `BackendWASM`。

- [ ] **步骤 3：创建 `pkg/zxing/wasm_stub.go`**

`//go:build cgo && (linux || windows) && !(js && wasm)`，提供 `wasmZXing` stub struct，所有方法返回 "WASM backend not available" 错误。

- [ ] **步骤 4：简化 `factory.go`**

`New(config)` 中 `BackendAuto`：编译期只有一个后端可用，直接返回。`NewCGO` 返回 `&cgoZXing{}`。`NewWASM` 返回 `&wasmZXing{}`。删除 `newAuto` 中的运行时检测逻辑。

- [ ] **步骤 5：简化 `universal_impl.go`**

删除 `universal_impl.go`，其逻辑已由 `cgo_impl.go` 和 `wasm_impl.go` 各自实现。`factory.go` 直接返回 `cgoZXing` 或 `wasmZXing`。

- [ ] **步骤 6：更新 `pkg/wasm/runtime_stub.go`**

build tag 改为 `//go:build cgo && (linux || windows) && !(js && wasm)`。

- [ ] **步骤 7：验证两种模式构建**

```bash
CGO_ENABLED=1 CGO_CFLAGS="-I$(pwd)/include" CGO_CXXFLAGS="-std=c++17 -I$(pwd)/include" \
CGO_LDFLAGS="-L$(pwd)/lib -lzxingwrapper -lZXing -lstdc++ -lm" go build ./pkg/zxing/
CGO_ENABLED=0 go build ./pkg/zxing/
```

- [ ] **步骤 8：Commit**

```bash
git add -A pkg/zxing/ pkg/wasm/
git commit -m "refactor: compile-time backend selection via build tags"
```

---

## 任务 5：创建构建工具 `cmd/build/`

**目标：** 纯 Go 跨平台构建工具，替代所有碎片化脚本。

**文件：** 创建 `cmd/build/main.go`, `env.go`, `env_test.go`, `build_lib.go`, `build_wasm.go`, `build_go.go`, `test.go`, `clean.go`, `docker_build.go`, `sync_headers.go`

- [ ] **步骤 1：创建 `env.go` — CGO 环境变量工具**

实现 `buildCGOEnv() []string`：返回包含 `CGO_ENABLED=1`、`CGO_CFLAGS`、`CGO_CXXFLAGS`、`CGO_LDFLAGS` 的 `[]string`（基于 `runtime.GOOS` 选择 `lib/{os}-{arch}` 路径）。使用 `os.Environ()` 作为基础，不污染当前进程。实现 `absPath(string) string` 辅助函数。实现 `detectArch() string` 返回 "x64" 或 "arm64"。

- [ ] **步骤 2：创建 `env_test.go`**

测试 `buildCGOEnv()` 返回的环境变量包含正确的 key、路径格式正确、不修改 `os.Environ()`。

- [ ] **步骤 3：创建 `main.go` — 入口和子命令分发**

使用 `flag` 包，第一个参数为子命令，其余为子命令参数。支持 `build-lib`, `build-wasm`, `build-go`, `build-all`, `sync-headers`, `test`, `clean`, `docker-build`。显示帮助信息。

- [ ] **步骤 4：创建 `build_lib.go`**

调用 `git submodule update --init --recursive`，然后调用 `cmake` 构建静态库到 `build/` 目录，最后复制 `.a`/`.lib` 文件到 `lib/{os}-{arch}/`。检测平台选择 CMake generator（Linux: Unix Makefiles，Windows: MinGW Makefiles）。

- [ ] **步骤 5：创建 `build_wasm.go`**

检查 `$EMSDK` 环境变量，备份 `CMakeLists.txt`，复制 `CMakeLists-wasm.txt` 为 `CMakeLists.txt`，调用 `emcmake cmake` + `emmake make`，复制产物到 `wasm/`，恢复原始 `CMakeLists.txt`。

- [ ] **步骤 6：创建 `build_go.go`**

设置 CGO 环境变量（通过 `exec.Cmd.Env`），运行 `go build ./...`。同时构建 `cmd/zxing-cli`。

- [ ] **步骤 7：创建 `test.go`**

设置 CGO 环境变量，运行 `go test ./pkg/... -v`。

- [ ] **步骤 8：创建 `clean.go`**

删除 `build/`、`build-wasm/`。不删除 `lib/` 和 `wasm/`（预编译产物）。

- [ ] **步骤 9：创建 `docker_build.go`**

构建 Docker 镜像并运行容器，挂载项目目录，在 CentOS 7 中编译 Linux 静态库。

- [ ] **步骤 10：创建 `sync_headers.go`**

从 `zxing-cpp/core/src/` 复制 `*.h` 到 `include/ZXing/`。确保 `include/zxing.h` 和 `include/zxing_internal.h` 在项目根目录的 `include/` 中。

- [ ] **步骤 11：运行 env 测试**

```bash
go test ./cmd/build/ -v -run TestEnv
```

- [ ] **步骤 12：验证构建工具可用**

```bash
go run ./cmd/build build-go
```

- [ ] **步骤 13：Commit**

```bash
git add cmd/build/
git commit -m "feat: add cross-platform build tool (cmd/build/)"
```

---

## 任务 6：创建 Docker 构建环境

**目标：** CentOS 7 Docker 镜像用于构建 glibc 2.17 兼容的 Linux 静态库。

**文件：** 创建 `docker/Dockerfile.linux-build`

- [ ] **步骤 1：创建 Dockerfile**

```dockerfile
FROM centos:7
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*.repo && \
    sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*.repo && \
    yum install -y centos-release-scl && \
    sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-SCLo.repo && \
    sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-SCLo.repo && \
    yum install -y devtoolset-7 cmake3 make git && \
    yum clean all
SHELL ["/usr/bin/scl", "enable", "devtoolset-7"]
WORKDIR /workspace
```

- [ ] **步骤 2：验证 Docker 构建**

```bash
go run ./cmd/build docker-build
```

- [ ] **步骤 3：Commit**

```bash
git add docker/Dockerfile.linux-build
git commit -m "feat: add CentOS 7 Docker build environment"
```

---

## 任务 7：构建预编译库并提交

**目标：** 生成 Linux 和 Windows 预编译静态库，创建 `BUILDINFO.md`，提交到仓库。

**文件：** `lib/linux-x64/*`, `lib/windows-x64/*`, `lib/BUILDINFO.md`

- [ ] **步骤 1：构建 Linux 静态库（Docker）**

```bash
go run ./cmd/build docker-build
```
产物：`lib/linux-x64/libZXing.a`, `lib/linux-x64/libzxingwrapper.a`

- [ ] **步骤 2：同步头文件**

```bash
go run ./cmd/build sync-headers
```

- [ ] **步骤 3：创建 `lib/BUILDINFO.md`**

记录：zxing-cpp v2.3.0 commit hash、构建日期、Docker 镜像、GCC 版本、SHA-256 校验和、构建命令。

- [ ] **步骤 4：验证 `go build` 直接可用（无需手动设 CGO 环境变量）**

```bash
CGO_ENABLED=1 go build ./pkg/zxing/
```
预期：通过（`#cgo` 指令自动指向 `lib/linux-x64`）

- [ ] **步骤 5：Commit**

```bash
git add lib/ include/
git commit -m "feat: add pre-compiled static libraries with BUILDINFO"
```

---

## 任务 8：删除旧脚本和文件

**目标：** 删除所有被 `cmd/build/` 替代的碎片化脚本。

- [ ] **步骤 1：删除旧构建脚本**

```bash
git rm build.sh build.bat build.ps1 build_all.ps1 build_wasm.ps1 build_wasm.sh build_wasm_demo.ps1
git rm scripts/build_static_linux.sh scripts/build_static_windows.ps1 scripts/build_wasm_save.sh scripts/build_wasm_save.ps1
git rm test_build.ps1 test_integration.ps1 test_cmake.sh
git rm Makefile
git rm -r lib64/
```

- [ ] **步骤 2：删除根目录 `zxing.go`**

`zxing.go` 是旧的 CGO 绑定（根包），已被 `pkg/zxing/` 完全替代。

```bash
git rm zxing.go
```

- [ ] **步骤 3：验证构建仍然通过**

```bash
CGO_ENABLED=1 go build ./pkg/zxing/ && CGO_ENABLED=0 go build ./pkg/zxing/
```

- [ ] **步骤 4：Commit**

```bash
git commit -m "refactor: remove fragmented build scripts, replaced by cmd/build/"
```

---

## 任务 9：更新测试

**目标：** 更新测试用例，验证两种后端可用。

**文件：** 修改 `pkg/zxing/zxing_test.go`

- [ ] **步骤 1：更新 `TestBackendSelection`**

改为验证编译期后端：CGO 模式下 `GetBackend()` 返回 `BackendCGO`，WASM 模式下返回 `BackendWASM`。使用 `//go:build` 条件编译或运行时判断 `CGO_ENABLED`。

- [ ] **步骤 2：更新 `TestDecodeBytes`**

使用 `data/qrcode_www.bing.com.png` 真实图片，加载为 `image.Image`，调用 `DecodeImage`，验证 `result.Text` 非空。

- [ ] **步骤 3：添加 `TestEncodeDecode`**

仅 WASM 后端：`EncodeText("Hello")` → `DecodeImage(result)` → 验证解码文本匹配。

- [ ] **步骤 4：运行测试（两种模式）**

```bash
CGO_ENABLED=1 go run ./cmd/build test
CGO_ENABLED=0 go test ./pkg/zxing/ -v
```

- [ ] **步骤 5：Commit**

```bash
git add pkg/zxing/zxing_test.go
git commit -m "test: update tests for dual-backend validation"
```

---

## 任务 10：更新 README.md

**目标：** 重写 README，反映新的构建和使用方式。

**文件：** 修改 `README.md`

- [ ] **步骤 1：重写 README**

按设计文档中的 README 结构重写：简介 → 快速开始（预编译库方式 + 从源码构建方式）→ 后端选择表 → 使用示例 → CLI 工具 → 构建工具命令 → 项目结构 → API 文档 → 测试 → 许可证。包含 Windows MinGW-w64 前提说明。包含 `git submodule update --init --recursive` 说明。

- [ ] **步骤 2：Commit**

```bash
git add README.md
git commit -m "docs: rewrite README for production-ready usage"
```

---

## 任务 11：最终验证

**目标：** 验证所有构建模式可用。

- [ ] **步骤 1：Linux CGO 构建**

```bash
CGO_ENABLED=1 go build ./...
```

- [ ] **步骤 2：Linux WASM 构建**

```bash
CGO_ENABLED=0 go build ./...
```

- [ ] **步骤 3：Linux CGO 测试**

```bash
CGO_ENABLED=1 go test ./pkg/... -v
```

- [ ] **步骤 4：Linux WASM 测试**

```bash
CGO_ENABLED=0 go test ./pkg/... -v
```

- [ ] **步骤 5：构建工具完整流程**

```bash
go run ./cmd/build build-all
```

- [ ] **步骤 6：更新 doc/requirements.md 和 doc/requirements-analysis.md**

追加本次需求和分析摘要。

- [ ] **步骤 7：最终 Commit**

```bash
git add doc/requirements.md doc/requirements-analysis.md
git commit -m "docs: update requirements and analysis for production-ready refactor"
```

---

## 自检

### 规格覆盖度

| 规格章节 | 覆盖任务 |
|----------|----------|
| Architecture / Package Structure | 任务 1, 4 |
| Backend Selection Matrix | 任务 1, 4 |
| CGO Backend Design | 任务 1, 7 |
| WASM Backend Design (wazero) | 任务 0, 2, 3, 4 |
| Build Tool cmd/build/ | 任务 5 |
| Docker build environment | 任务 6 |
| Pre-compiled libraries | 任务 7 |
| Files to Delete | 任务 8 |
| Testing Strategy | 任务 9 |
| README Structure | 任务 10 |
| API Backward Compatibility | 任务 4 (factory.go 保持 API 不变) |
| Error Messages | 任务 5 (build-go 检测缺失库) |
| Implementation Order | 任务 0-11 顺序匹配 |

### 占位符扫描

无 TODO/待定/占位符。所有步骤包含具体代码描述或命令。

### 类型一致性

- `BarcodeFormat` 在 `cgo_binding_linux.go`、`cgo_binding_windows.go`、`cgo_stub.go` 中定义一致
- `CGODecodeOptions`/`CGODecodeResult` 在三个文件中结构一致
- `wasm.Runtime` 的 `DecodeResult`/`EncodeResult` 在 `runtime_wazero.go` 中定义，与 `runtime_js.go` 一致
- `cgoZXing` 和 `wasmZXing` 都实现 `ZXing` interface（`DecodeImage`/`DecodeBytes`/`EncodeText`/`EncodeToBytes`/`Close`/`GetBackend`）
- `factory.go` 的 `New()` 返回 `ZXing` interface
