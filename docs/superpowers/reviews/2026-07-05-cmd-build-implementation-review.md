# Code Review: `cmd/build/build_go.go` 及构建工具集成

## 评审对象

- **主要文件**: `cmd/build/build_go.go`
- **关联文件**: `cmd/build/env.go`, `cmd/build/main.go`, `cmd/build/test.go`, `cmd/build/build_lib.go`, `cmd/build/wasm_build.go`
- **日期**: 2026-07-05
- **评审结论**: 结构合理，但 `build_go.go` 对 CGO/非 CGO 的自动检测逻辑与规格说明存在偏差，需要修正。

---

## 总体评价

`cmd/build` 模块整体进展良好，`env.go` 已正确实现“不污染 `os.Environ()`”的环境变量构造，并配套了单元测试。`main.go` 的命令结构清晰。

本次新增的 `build_go.go` 完成了 `build-go` 与 `build-all` 两个命令，但存在以下关键问题：

1. **未检测预编译库是否存在**；
2. **未尊重用户传入的 `CGO_ENABLED` 偏好**；
3. **CGO 失败回退逻辑过于狭窄**（仅在 `projectRoot()` 失败时才回退）。

这三个问题导致 `build-go` 的行为与规格说明中的“自动检测预编译库，设置 CGO 环境变量”不符。

---

## 优点

- **不污染进程环境**: `buildCGOEnv()` 通过 `cmd.Env = env` 注入子进程环境，且 `env_test.go` 验证 `os.Environ()` 不被修改，符合生产就绪要求。
- **命令结构一致**: `build_go.go` 的函数签名与 `build_lib.go`、`build_wasm.go` 保持一致，便于维护。
- **与 `test.go` 共享逻辑**: `runTest` 也复用了 `buildCGOEnv()` / `buildNonCGOEnv()`，逻辑统一。

---

## 必须修正的问题

### 1. `buildGo` 未检测预编译库是否存在

`@/data/dev/github.com/chennqqi/zxing/cmd/build/build_go.go:11-15`

```go
env, err := buildCGOEnv()
if err != nil {
    fmt.Printf("Warning: CGO env setup failed (%v), using non-CGO build\n", err)
    env = buildNonCGOEnv()
}
```

`buildCGOEnv()` 仅在 `projectRoot()` 失败时返回错误。即使 `lib/{linux,windows}-x64/` 中完全没有静态库，它也会返回 CGO 环境变量，导致 `go build` 在链接阶段失败，报错信息晦涩。

**建议**: 在 `buildCGOEnv()` 调用前或内部检查库文件是否存在。例如：

```go
func hasPrebuiltLibs() bool {
    libPath := libDir()
    // 检查 libzxingwrapper.a / libZXing.a 或 windows 对应 .lib
    ...
}
```

如果库不存在，则直接使用 `buildNonCGOEnv()`。

### 2. 未尊重用户传入的 `CGO_ENABLED` 偏好

`@/data/dev/github.com/chennqqi/zxing/cmd/build/build_go.go:11`

当前 `buildGo` 总是优先尝试 CGO，忽略用户可能显式设置的 `CGO_ENABLED=0`。这会导致以下场景失败：

- 用户已安装 gcc，但只想要 WASM 后端测试；
- 预编译库存在但用户希望验证非 CGO 构建。

**建议**: 先读取 `os.Getenv("CGO_ENABLED")`：

- `"0"` → 强制非 CGO；
- `"1"` → 强制 CGO，并检查库是否存在；
- 空 → 自动检测库是否存在，决定使用 CGO 还是非 CGO。

### 3. 非 CGO 回退时机不正确

`@/data/dev/github.com/chennqqi/zxing/cmd/build/build_go.go:12-15`

当前只有在 `buildCGOEnv()` 返回错误时才回退。如前所述，`buildCGOEnv()` 几乎不会失败。因此“CGO env setup failed, using non-CGO build”这条警告在生产环境中几乎不会触发，而库缺失时会直接报错。

**建议**: 将库缺失作为非 CGO 回退的条件，而不是仅依赖 `buildCGOEnv()` 的错误。

### 4. `buildAll` 顺序固定，缺乏跳过机制

`@/data/dev/github.com/chennqqi/zxing/cmd/build/build_go.go:41-52`

```go
func buildAll(args []string) error {
    if err := buildLib(args); err != nil {
        return fmt.Errorf("build-lib failed: %w", err)
    }
    if err := buildWasm(args); err != nil {
        return fmt.Errorf("build-wasm failed: %w", err)
    }
    if err := buildGo(args); err != nil {
        return fmt.Errorf("build-go failed: %w", err)
    }
    ...
}
```

如果本地没有 CMake 或 EMSDK，`build-all` 会直接失败。用户即使只想验证已有预编译库下的 Go 构建，也无法跳过。

**建议**: 增加 `build-all` 的 `-skip-lib` / `-skip-wasm` 标志，或至少让 `build-all` 在缺少依赖时给出友好提示，而不是直接失败。也可以考虑让 `build-all` 接受 `args` 并传递给子命令。

---

## 建议优化

### 1. 支持 Windows 可执行文件扩展名

`@/data/dev/github.com/chennqqi/zxing/cmd/build/build_go.go:28`

```go
cmd = exec.Command("go", "build", "-o", "bin/zxing-cli", "./cmd/zxing-cli")
```

在 Windows 上应输出 `bin/zxing-cli.exe`，否则生成的可执行文件无法直接运行。

**建议**: 使用 `runtime.GOOS` 判断：

```go
outputPath := "bin/zxing-cli"
if runtime.GOOS == "windows" {
    outputPath += ".exe"
}
```

### 2. `buildGo` 应透传 `args`

`@/data/dev/github.com/chennqqi/zxing/cmd/build/build_go.go:10`

函数签名接收 `args []string`，但内部完全忽略。用户可能希望传入 `go build` 的额外标志（如 `-v`、`-x`、特定包路径）。

**建议**: 将 `args` 追加到 `go build` 命令中，例如：

```go
cmdArgs := append([]string{"build"}, args...)
cmdArgs = append(cmdArgs, "./...")
cmd := exec.Command("go", cmdArgs...)
```

同时注意 `build-all` 的 `args` 透传策略，避免将所有标志同时传给 `build-lib`、`build-wasm`、`build-go`。

### 3. 对 `go build ./...` 与 `go build ./cmd/zxing-cli` 的解释

当前 `build-go` 会顺序执行两次构建：一次是整个项目，一次是 CLI。这可能导致重复编译。

**建议**: 如果 `go build ./...` 已经包含 `cmd/zxing-cli`，第二次构建可以省略，或改为显式构建 CLI 作为可交付产物。如果保留两次构建，应在文档中说明原因。

### 4. `test.go` 与 `build_go.go` 行为一致性

`@/data/dev/github.com/chennqqi/zxing/cmd/build/test.go:10-15`

`runTest` 与 `buildGo` 使用了相同的 CGO 回退逻辑。当 `buildGo` 修正后，应将公共逻辑提取到 `env.go` 中的函数，例如 `selectBuildEnv()`，避免两处重复。

### 5. 缺失的 `build-go` 行为说明

建议更新 `main.go` 的 `usageText`，说明 `build-go` 会优先尝试 CGO，并允许通过 `CGO_ENABLED=0` 强制非 CGO。

---

## 验证清单

`build_go.go` 修正后应通过以下检查：

- [ ] 在缺少 `lib/linux-x64/*.a` 时，`build-go` 自动回退到非 CGO 并构建成功
- [ ] 在存在 `lib/linux-x64/*.a` 时，`build-go` 使用 CGO 构建成功
- [ ] `CGO_ENABLED=0 go run ./cmd/build build-go` 强制使用非 CGO 构建
- [ ] `CGO_ENABLED=1 go run ./cmd/build build-go` 在缺少库时给出明确错误，而不是链接失败
- [ ] Windows 上 `build-go` 输出 `bin/zxing-cli.exe`
- [ ] `go run ./cmd/build test` 在 CGO 与非 CGO 模式下均通过
- [ ] `build-all` 在缺少 EMSDK 或 CMake 时给出友好提示，而非直接崩溃

---

## 结论

`cmd/build` 整体方向正确，新增的 `build_go.go` 完成了命令骨架。但**必须修正 CGO 自动检测逻辑**，否则 `build-go` 会在预编译库缺失时直接失败，而不是按规格自动回退到 WASM 后端。修正后，该工具可以进入测试阶段。
