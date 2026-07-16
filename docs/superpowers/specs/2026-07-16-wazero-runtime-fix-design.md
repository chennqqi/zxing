# Wazero Runtime Fix Design

## Goal

Make the server-side wazero backend reliable for production decoding while preserving the public `pkg/zxing` API and existing build-tag selection.

The fix must eliminate the current out-of-bounds memory failure, make lifecycle and decode calls safe under concurrency, propagate cancellation, clean up all wazero resources on failure, and apply public decode options consistently.

WASM barcode encoding is outside this change. It will continue to return an explicit unsupported error.

## Root Cause and Constraints

The current data path converts raw pixels to PNG in Go, copies the PNG into a guest-global 16 MiB bump buffer, and decodes it again with `stb_image` inside WASM. The checked-in module fails during this path with an out-of-bounds memory access. The bump allocator also wraps without rejecting oversized allocations and is reset before every call, so concurrent calls can overwrite live guest memory.

The module contains additional shared state, including its allocator and last-error buffer. A single module instance therefore cannot execute independent decode calls concurrently.

The implementation must remain compatible with the repository's pure-Go fallback and must not change the Linux glibc static-library build process. The wazero dependency is upgraded from v1.8.0 to v1.12.0 as part of this work.

## Architecture

### Raw-pixel WASM ABI

Add a C ABI function that accepts a pointer to raw pixel bytes, width, height, channel count, and decode options. It constructs a ZXing `ImageView` directly and returns the existing `DecodeResult` structure.

Go allocates guest memory with the exported standard allocator, writes the pixels, invokes the raw-pixel function, copies the result into Go-owned memory, and releases every guest allocation before unlocking the module.

The custom static bump buffer and its reset function are removed from the wazero data path. This removes PNG encoding, `stb_image` decoding, the fixed 16 MiB input limit, and allocator wraparound.

### Lifecycle and concurrency

`Runtime` owns a mutex protecting initialization, guest calls, readiness checks, and close. A decode call holds this lock for the complete guest-memory transaction because guest memory and C++ globals are shared.

Initialization creates the wazero runtime with context-aware termination enabled. It uses error-returning WASI initialization and records the compiled module so that explicit cleanup is possible. Every partial-failure path closes resources before returning.

The implementation uses the non-deprecated wazero v1.12.0 APIs. `go.mod` records wazero as a direct dependency, and `go mod tidy` updates module metadata after the upgrade.

If cancellation closes a module, the runtime is marked unusable and its wazero resources are released. The public backend's lazy initializer checks readiness as well as a nil pointer, so a later operation initializes a fresh runtime rather than calling a closed module.

`Close` is idempotent and returns the underlying close error. It clears all stored state regardless of the close result.

### Context propagation

Runtime decode methods accept the caller's `context.Context` and use it for every guest function call. They do not replace it with `context.Background`.

Cancellation or deadline expiration returns a wrapped error to the public API. Since wazero cancellation can close the module, the runtime clears its ready state when that occurs.

### Decode options

The Go runtime accepts the public decode settings needed by the guest and writes all fields of the C `DecodeOptions` structure using an exported configuration function or an ABI-safe setter function. Direct dependence on undocumented struct offsets is avoided.

The C++ raw-pixel decode function applies formats, try-harder, rotation, inversion, and downscaling consistently with the native decode path.

### Validation and errors

Go validates non-empty input, positive dimensions, supported channel counts, integer overflow, and exact minimum buffer length before allocating guest memory.

All required exports are resolved and validated during initialization. Missing exports fail initialization instead of causing a later nil-function panic.

Guest error strings are read with bounded memory access. Errors from allocation, memory writes, calls, and cleanup are wrapped with the operation name.

## Testing

Development follows red-green TDD. Tests are added before production changes and must fail for the intended missing behavior.

Coverage includes:

- Loading the checked-in WASM module and decoding the existing QR fixture.
- Passing public decode options through the WASM backend.
- Rejecting empty, undersized, overflowed, and invalid-dimension buffers.
- Concurrent decode calls on one public ZXing instance.
- Context cancellation or deadline termination.
- Initialization failure followed by cleanup.
- Repeated and concurrent `Close` calls.
- Detection of missing required exports where practical.
- Compatibility with wazero v1.12.0 and a tidy module graph.

Verification commands are:

```powershell
$env:CGO_ENABLED='0'; go test ./... -count=1
go test ./... -count=1
$env:CGO_ENABLED='0'; go test -race ./pkg/wasm ./pkg/zxing -count=1
```

The WASM module is rebuilt with the repository's supported Emscripten workflow after the ABI change and the resulting checked-in artifact is tested, not only a locally substituted module.

## Non-goals

- Implementing barcode encoding in the wazero backend.
- Introducing a module pool or parallel guest execution.
- Changing CGO backend selection or supported platforms.
- Changing the Linux static-library build environment.
