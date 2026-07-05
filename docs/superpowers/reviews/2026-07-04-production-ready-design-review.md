# Design Review: Production-Ready zxing Go Library

## 评审对象

- **文档**: `docs/superpowers/specs/2026-07-04-production-ready-design.md`
- **日期**: 2026-07-04
- **评审结论**: 方案整体可行，方向正确，但存在若干**必须修正**的构建标签与跨平台问题。建议在进入阶段 2 之前先澄清并修复这些问题。

---

## 总体评价

该方案提出了一套清晰的“预编译静态库 + wazero WASM 运行时 + 统一 Go 构建工具”的改造思路，核心目标是：

1. 统一 Windows / Linux 构建流程
2. 消除运行时后端切换，改用编译期构建标签
3. 将 WASM 后端从浏览器环境扩展到服务端 Go
4. 用纯 Go 构建工具替换大量 sh / ps1 / bat 脚本

整体架构合理，但需要对构建标签矩阵、Windows CGO 链接方式、预编译二进制库管理做进一步澄清和约束。

---

## 优点

- **统一构建入口**: 使用 `cmd/build` 替代碎片化脚本，降低维护成本，跨平台体验一致。
- **编译期后端选择**: 去掉 `universal_impl.go` 的运行时切换，减少运行时复杂度和意外分支。
- **服务端 WASM 支持**: 引入 wazero 让 `CGO_ENABLED=0` 也能在服务端运行，扩展了使用场景。
- **明确的库目录结构**: `lib/{linux-x64,windows-x64}` 与 `include/` 的分离方式清晰。
- **清晰的文件删除清单**: 哪些旧脚本由新工具替代，罗列清楚。

---

## 必须修正 / 高度关注的问题

### 1. 构建标签矩阵存在平台覆盖缺口

`pkg/zxing/` 中提议的标签如下：

- `cgo_impl.go` : `//go:build cgo`
- `cgo_binding.go` : `//go:build cgo && linux`
- `cgo_binding_windows.go` : `//go:build cgo && windows`
- `cgo_stub.go` : `//go:build !cgo`
- `wasm_impl.go` : `//go:build !cgo`
- `wasm_impl_js.go` : `//go:build js && wasm`
- `wasm_stub.go` : `//go:build cgo && !(js && wasm)`

**问题**: 当 `CGO_ENABLED=1` 且目标平台为 `darwin` 或 `freebsd` 时，`cgo_impl.go` 会生效，但 `cgo_binding.go` 和 `cgo_binding_windows.go` 都不会生效，导致缺少 `#cgo` 指令，编译失败。

**建议**:

```go
// cgo_impl.go
//go:build cgo && (linux || windows)
```

或在 `cgo_impl.go` 中保留默认的跨平台 `#cgo` 配置。文档应明确声明仅支持 `linux` 与 `windows` 的 CGO 后端。

### 2. Windows CGO 与 MSVC 静态库存在根本性冲突

方案计划：

- 静态库在 Windows 下使用 **MSVC 2022** 编译
- CGO 后端通过 `#cgo LDFLAGS` 链接这些库

**问题**: Go 的 CGO 在 Windows 上默认使用 **MinGW-w64 GCC**，而不是 MSVC。MSVC 编译的 `.lib` 使用 COFF / MS 名称修饰，MinGW GCC 通常无法直接链接 MSVC C++ 静态库，尤其是涉及 C++ ABI、异常处理和 STL 时。

**建议**（择一）:

1. Windows 静态库也改用 **MinGW-w64** 编译，与 CGO 工具链一致；
2. 或明确说明 Windows 下 CGO 需要 MSVC 兼容的 `gcc`（如 `llvm-mingw`），并提供经过验证的链接方式；
3. 在 `cmd/build` 中检测 Windows 编译器环境，给出明确错误提示，避免用户直接 `go build` 失败。

### 3. CentOS 7 已 EOL

方案使用 CentOS 7 构建 Linux 静态库以获得 glibc 2.17 兼容性。CentOS 7 已于 2024 年 6 月 EOL，官方仓库可能不再可用。

**建议**:

- 使用 `centos:7` 镜像时需配置 Vault 源，或使用 AlmaLinux 7 / Rocky Linux 7 的兼容构建环境；
- 在 `docker/Dockerfile.linux-build` 中显式声明并测试仓库可用性；
- 文档中说明兼容性目标（glibc 2.17）以及 EOL 后的镜像维护策略。

### 4. 预编译二进制库进入版本库的可审计性

方案计划将 `lib/linux-x64/*.a`、`lib/windows-x64/*.lib` 和 `wasm/zxingwrapper.wasm` 提交到仓库。

**问题**:

- 二进制文件无法通过 diff 审计，存在供应链安全风险；
- 仓库体积会显著膨胀；
- 用户无法确认这些库与源码是否一致。

**建议**:

- 在 `lib/` 下增加 `BUILDINFO.md`，记录库版本、源码 commit hash、构建时间、编译器版本、校验和；
- 考虑使用 Git LFS 管理大二进制文件；
- 在 CI 中增加“从源码重新构建并比对校验和”的回归检查。

### 5. `cmd/build` 的 `setupCGOEnv` 影响当前进程环境

```go
os.Setenv("CGO_CFLAGS", ...)
os.Setenv("CGO_CXXFLAGS", ...)
os.Setenv("CGO_LDFLAGS", ...)
os.Setenv("CGO_ENABLED", "1")
```

**问题**: 修改当前进程环境变量，在并发调用或作为库被调用时可能产生副作用。虽然 `cmd/build` 是命令行工具，但仍建议显式构建 `exec.Command` 的 `Env` 字段，避免污染。

---

## 建议优化

1. **构建工具自身可测试化**: 将 `cmd/build` 中的平台检测、环境设置、路径拼接逻辑抽出为可测试函数，并编写单元测试。
2. **向后兼容说明**: 当前项目已有 `zxing.go`、`universal_impl.go` 等文件，重构后 API 是否完全保持不变需要说明。如果 `New()` 等函数签名变化，需给出迁移示例。
3. **WASM 运行时生命周期管理**: wazero 的 `Runtime` 和 `Module` 建议提供 `Close()` 方法，避免 goroutine 和内存泄漏。
4. **错误信息增强**: 当用户在没有预编译库的环境中运行 `go build` 时，应给出清晰提示，例如：
   - 缺少哪个平台/架构的库；
   - 如何运行 `cmd/build build-lib` 或 `docker-build` 重新生成。
5. **README 明确 Windows 前提**: 在快速开始部分说明 Windows 用户必须安装 MinGW-w64 或兼容的 gcc 工具链才能使用 CGO 后端。
6. **CI 矩阵覆盖**: 建议在 CI 中不仅测试 `CGO_ENABLED=1/0`，还要测试 `GOOS=js GOARCH=wasm` 的构建，以及从源码重新构建静态库。

---

## 风险与依赖

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Windows CGO 链接 MSVC 静态库失败 | 高 | 统一使用 MinGW 工具链，或验证 MSVC 兼容方案 |
| CentOS 7 EOL 导致 Docker 构建失败 | 中 | 改用 Vault 源或 AlmaLinux 7 兼容镜像 |
| 预编译库与源码不一致 | 中 | 增加校验和、BUILDINFO、CI 回归检查 |
| wazero WASM 性能不达预期 | 中 | 性能基准测试，与 CGO 对比 |
| 构建标签矩阵遗漏平台 | 中 | 明确限制支持平台，增加 `go vet` 或 CI 多平台构建验证 |

---

## 验证清单（重构完成后应执行）

- [ ] `CGO_ENABLED=1 go build ./...` 在 Linux 上通过
- [ ] `CGO_ENABLED=0 go build ./...` 在 Linux 上通过（使用 wazero）
- [ ] Windows 上 `CGO_ENABLED=1 go build ./...` 通过（需确认编译器工具链）
- [ ] Windows 上 `CGO_ENABLED=0 go build ./...` 通过
- [ ] `GOOS=js GOARCH=wasm go build ./...` 通过
- [ ] `go run ./cmd/build build-all` 在 Linux 和 Windows 上工作
- [ ] `go run ./cmd/build docker-build` 成功生成 Linux 静态库
- [ ] 生成的 Linux 静态库在 CentOS 7 / 8 / Ubuntu 16.04+ 上可链接
- [ ] 使用真实二维码图片进行 CGO 与 WASM 解码结果对比测试
- [ ] 新增 `TestEncodeDecode` 在 WASM 后端下通过
- [ ] README 中的快速开始命令可被新用户按步骤执行成功

---

## 结论

方案在架构层面是合理且值得执行的，尤其解决了构建脚本碎片化和 WASM 服务端支持的问题。但 **Windows CGO 与 MSVC 静态库的兼容性**、**构建标签的平台覆盖**、**预编译二进制库的可审计性** 是重构开始前必须解决的阻塞性问题。建议在进入大规模文件迁移前，先完成一个最小可行的 PoC：

1. 在 Windows 上验证 CGO 能否成功链接 MSVC 或 MinGW 编译的静态库；
2. 在 Linux 上验证 `CGO_ENABLED=0` 时 wazero 能否成功解码二维码；
3. 确定最终目录结构和文件删除清单后，再批量删除旧脚本。
