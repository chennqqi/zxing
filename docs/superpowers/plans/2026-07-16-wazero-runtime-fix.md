# Wazero Runtime Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a production-safe wazero backend that decodes raw pixels without out-of-bounds failures, honors cancellation and decode options, and uses wazero v1.12.0.

**Architecture:** Add a raw-pixel C ABI and rebuild the checked-in WASM module. Serialize each module instance's lifecycle and guest-memory transactions, propagate caller contexts into wazero, validate inputs before allocation, and map options through ABI functions instead of struct offsets.

**Tech Stack:** Go 1.24, wazero v1.12.0, C++20, zxing-cpp, Emscripten standalone WASM, Go test/race detector.

## Global Constraints

- Preserve the public `pkg/zxing` API and build-tag selection.
- Use wazero v1.12.0 as a direct dependency and leave the module graph tidy.
- Do not change the CentOS 7/glibc 2.17 Linux library build.
- Do not implement WASM encoding or introduce a module pool.
- New functions require English doc comments.
- Preserve unrelated user state, including `_tmp_commit_diff.txt`.

## File Map

- `include/zxing.h`, `src/zxing.cpp`: raw-pixel and option ABI.
- `CMakeLists-wasm.txt`, `wasm/zxingwrapper.wasm`: exports and rebuilt artifact.
- `pkg/wasm/runtime_wazero.go`: lifecycle, synchronization, validation, and guest calls.
- `pkg/wasm/runtime_wazero_test.go`: ABI, decode, lifecycle, cancellation, and concurrency tests.
- `pkg/zxing/wasm_impl.go`, `pkg/zxing/zxing_test.go`: public option/context integration.
- `go.mod`, `go.sum`: wazero v1.12.0.

---

### Task 1: Upgrade wazero and lock in the failing regression

**Files:**
- Modify: `go.mod`, `go.sum`
- Modify: `pkg/wasm/runtime_wazero_test.go`

**Interfaces:**
- Consumes: existing `NewRuntime`, `Initialize`, and `DecodeImage`.
- Produces: a direct wazero v1.12.0 dependency and exact QR regression assertion.

- [ ] **Step 1: Strengthen the existing QR test**

```go
if result.Text != "https://www.bing.com/" {
	t.Fatalf("unexpected decoded text: %q", result.Text)
}
if result.Format != "QR_CODE" {
	t.Fatalf("unexpected decoded format: %q", result.Format)
}
```

- [ ] **Step 2: Verify RED**

Run: `$env:CGO_ENABLED='0'; go test ./pkg/wasm -run TestWazeroDecodeQRCode -count=1 -v`

Expected: FAIL with `out of bounds memory access`.

- [ ] **Step 3: Upgrade and tidy**

```powershell
go get github.com/tetratelabs/wazero@v1.12.0
go mod tidy
go list -m github.com/tetratelabs/wazero
```

Expected module output: `github.com/tetratelabs/wazero v1.12.0`. Re-run Step 2 and expect the same OOB, proving the version update alone is not the functional fix.

- [ ] **Step 4: Commit**

```powershell
git add go.mod go.sum pkg/wasm/runtime_wazero_test.go
git commit -m "build: upgrade wazero to v1.12.0"
```

### Task 2: Add and rebuild the raw-pixel WASM ABI

**Files:**
- Modify: `include/zxing.h`, `src/zxing.cpp`, `CMakeLists-wasm.txt`
- Modify: `wasm/zxingwrapper.wasm`
- Test: `pkg/wasm/runtime_wazero_test.go`

**Interfaces:**
- Consumes: existing `DecodeOptions`, `DecodeResult`, `free_result`.
- Produces: `configure_decode_options` and `decode_barcode_pixels` exports.

- [ ] **Step 1: Add an export-presence test**

```go
func TestWazeroRequiredExports(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	for _, name := range []string{"malloc", "free", "configure_decode_options", "decode_barcode_pixels"} {
		if rt.module.ExportedFunction(name) == nil {
			t.Fatalf("required export %q is missing", name)
		}
	}
}
```

Run: `$env:CGO_ENABLED='0'; go test ./pkg/wasm -run TestWazeroRequiredExports -count=1 -v`

Expected: FAIL because the two new exports are absent.

- [ ] **Step 2: Declare the ABI**

```cpp
// Configures all fields of an existing decode options structure.
void configure_decode_options(DecodeOptions* options, int formats, int try_harder,
                              int try_rotate, int try_invert, int try_downscale);

// Decodes tightly packed raw pixels without an intermediate encoded image.
DecodeResult* decode_barcode_pixels(const unsigned char* data, int width, int height,
                                    int channels, const DecodeOptions* options);
```

- [ ] **Step 3: Implement the ABI**

In `src/zxing.cpp`, map channels 1/2/3/4 to `Lum/LumA/RGB/RGBA`, construct `ImageView` directly, and apply `formats`, `try_harder`, `try_rotate`, `try_invert`, and `try_downscale` to `ReaderOptions`. Reject null data, non-positive dimensions, unsupported channels, and allocation failures via `set_error`. Return the existing 12-byte wasm32 `DecodeResult`.

Use these exported definitions:

```cpp
EXPORT void configure_decode_options(DecodeOptions*, int, int, int, int, int);
EXPORT DecodeResult* decode_barcode_pixels(const unsigned char*, int, int, int,
                                           const DecodeOptions*);
```

Remove `zxing_alloc_buffer`, `zxing_alloc_offset`, `zxing_alloc`, and `zxing_alloc_reset`.

- [ ] **Step 4: Export and rebuild**

Replace custom allocator exports in `CMakeLists-wasm.txt` with the two new functions, retaining `malloc` and `free`.

Run: `go run ./cmd/build build-wasm`

Expected: `WASM build complete.` and a changed `wasm/zxingwrapper.wasm`.

- [ ] **Step 5: Verify GREEN and commit**

Run: `$env:CGO_ENABLED='0'; go test ./pkg/wasm -run TestWazeroRequiredExports -count=1 -v`

Expected: PASS.

```powershell
git add include/zxing.h src/zxing.cpp CMakeLists-wasm.txt wasm/zxingwrapper.wasm pkg/wasm/runtime_wazero_test.go
git commit -m "feat: add raw pixel wasm decode ABI"
```

### Task 3: Replace the Go host data path and validate inputs

**Files:**
- Modify: `pkg/wasm/runtime_wazero.go`
- Modify: `pkg/wasm/runtime_wazero_test.go`

**Interfaces:**
- Produces:

```go
type DecodeOptions struct {
	Formats      int
	TryHarder    bool
	TryRotate    bool
	TryInvert    bool
	TryDownscale bool
}

func (r *Runtime) DecodeImage(ctx context.Context, data []byte, width, height, channels int, opts *DecodeOptions) (*DecodeResult, error)
```

- [ ] **Step 1: Write validation tests**

Table cases: empty data, zero width, zero height, channel count 0 and 5, a three-byte buffer for one RGBA pixel, and `width=math.MaxInt, height=2, channels=4`. Each must return an error without entering WASM.

Run: `$env:CGO_ENABLED='0'; go test ./pkg/wasm -run TestDecodeImageRejectsInvalidInput -count=1 -v`

Expected: compile FAIL because the context/options signature does not exist.

- [ ] **Step 2: Implement minimal raw-pixel calls**

Validate positive dimensions, supported channels, multiplication overflow by division, and `len(data) >= width*height*channels`. Call standard guest `malloc`, write only required pixels, call `configure_decode_options`, call `decode_barcode_pixels`, copy the result/string into Go memory, then release result/options/input on every allocated path. Remove PNG encoding imports and helpers.

- [ ] **Step 3: Verify GREEN**

Run: `$env:CGO_ENABLED='0'; go test ./pkg/wasm -run "TestDecodeImageRejectsInvalidInput|TestWazeroDecodeQRCode" -count=1 -v`

Expected: PASS.

- [ ] **Step 4: Commit**

```powershell
git add pkg/wasm/runtime_wazero.go pkg/wasm/runtime_wazero_test.go
git commit -m "fix: decode raw pixels with wazero"
```

### Task 4: Make lifecycle, cancellation, and concurrency safe

**Files:**
- Modify: `pkg/wasm/runtime_wazero.go`
- Modify: `pkg/wasm/runtime_wazero_test.go`

**Interfaces:**
- Consumes: Task 3 API.
- Produces: synchronized initialization, calls, readiness, and idempotent close.

- [ ] **Step 1: Add failing lifecycle tests**

Add:

```go
func TestRuntimeConcurrentDecode(t *testing.T)
func TestRuntimeCloseIsConcurrentAndIdempotent(t *testing.T)
func TestDecodeImageHonorsCanceledContext(t *testing.T)
func TestInitializeFailureLeavesRuntimeNotReady(t *testing.T)
```

Start eight decode goroutines on one runtime and verify exact QR results. Start four concurrent closes and require no panic. Pass an already-canceled context and require a cancellation/exit error. Initialize a nonexistent path and require `!IsReady()`.

Run: `$env:CGO_ENABLED='0'; go test ./pkg/wasm -run "TestRuntime|TestDecodeImageHonors|TestInitializeFailure" -count=1 -v`

Expected: FAIL or race/panic under current shared state.

- [ ] **Step 2: Implement lifecycle safety**

Add a `sync.Mutex`, `wazero.CompiledModule`, and cached `api.Function` exports. Create the runtime with:

```go
config := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
runtime := wazero.NewRuntimeWithConfig(ctx, config)
```

Use `wasi_snapshot_preview1.Instantiate`, not `MustInstantiate`. Resolve all exports before setting ready. On partial failure close compiled module/runtime and clear fields. Hold the mutex for each complete guest transaction. `IsReady` checks `module != nil && !module.IsClosed()`. `Close` clears state and returns the first cleanup error.

- [ ] **Step 3: Verify GREEN and commit**

Run: `$env:CGO_ENABLED='0'; go test -race ./pkg/wasm -count=1`

Expected: PASS with no race report.

```powershell
git add pkg/wasm/runtime_wazero.go pkg/wasm/runtime_wazero_test.go
git commit -m "fix: synchronize wazero runtime lifecycle"
```

### Task 5: Integrate options and recovery in the public backend

**Files:**
- Modify: `pkg/zxing/wasm_impl.go`
- Modify: `pkg/zxing/zxing_test.go`

**Interfaces:**
- Consumes: `wasm.DecodeOptions` and context-aware `DecodeImage`.
- Produces: public option mapping, context propagation, and safe lazy reinitialization.

- [ ] **Step 1: Add failing public tests**

Test `PossibleFormats: []string{"QR_CODE"}` succeeds, an incompatible known format fails, an already-canceled context returns an error, and eight concurrent decodes on one `ZXing` instance return the exact QR value.

Run: `$env:CGO_ENABLED='0'; go test ./pkg/zxing -run "TestWASM.*(Options|Canceled|Concurrent)" -count=1 -v`

Expected: FAIL because options/context are discarded and initialization is unsynchronized.

- [ ] **Step 2: Implement mapping and recovery**

Protect `wasmZXing.runtime` with a mutex. Initialize when nil or `!IsReady()`; close/discard failed instances. Map `PossibleFormats` to existing bit flags and pass:

```go
&wasm.DecodeOptions{
	Formats:      formatFlags,
	TryHarder:    opts.TryHarder,
	TryRotate:    true,
	TryInvert:    false,
	TryDownscale: true,
}
```

Call `w.runtime.DecodeImage(ctx, data, width, height, 4, runtimeOpts)`. Detach the pointer under lock before `Close`.

- [ ] **Step 3: Verify GREEN and commit**

Run: `$env:CGO_ENABLED='0'; go test -race ./pkg/zxing -count=1`

Expected: PASS with no race report.

```powershell
git add pkg/zxing/wasm_impl.go pkg/zxing/zxing_test.go
git commit -m "fix: propagate wasm decode options and context"
```

### Task 6: Full verification and documentation alignment

**Files:**
- Modify if contradictory: `README.md`, `wasm/README.md`

**Interfaces:**
- Consumes: all prior tasks.
- Produces: verified builds/tests and accurate user documentation.

- [ ] **Step 1: Scan stale claims**

Run: `rg -n "wazero v1\.8|zxing_alloc|PoC: encode|WASM.*encod" README.md wasm pkg CMakeLists-wasm.txt src include`

Update only current user-facing contradictions; do not rewrite historical documents.

- [ ] **Step 2: Format and validate metadata**

```powershell
gofmt -w pkg/wasm/runtime_wazero.go pkg/wasm/runtime_wazero_test.go pkg/zxing/wasm_impl.go pkg/zxing/zxing_test.go
go mod tidy
go list -m github.com/tetratelabs/wazero
git diff --check
```

Expected: wazero v1.12.0 and no formatting/whitespace errors.

- [ ] **Step 3: Verify all relevant configurations**

```powershell
$env:CGO_ENABLED='0'; go test ./... -count=1
$env:CGO_ENABLED='0'; go build -buildvcs=false ./cmd/zxing-cli
go test ./... -count=1
$env:CGO_ENABLED='0'; go test -race ./pkg/wasm ./pkg/zxing -count=1
```

Expected: every command exits 0 and race tests report no races.

- [ ] **Step 4: Final review and optional docs commit**

```powershell
git status --short
git diff --check
git diff --stat HEAD~5..HEAD
```

If README files changed, commit only them:

```powershell
git add README.md wasm/README.md
git commit -m "docs: align wazero backend behavior"
```
