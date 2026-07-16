//go:build (!cgo || !(linux || windows)) && !(js && wasm)

// Package wasm provides ZXing WebAssembly runtime support via wazero.
package wasm

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// DecodeOptions controls barcode decoding behavior in the WASM backend.
type DecodeOptions struct {
	Formats      int
	TryHarder    bool
	TryRotate    bool
	TryInvert    bool
	TryDownscale bool
}

// Runtime manages the wazero WASM runtime for ZXing.
// A single Runtime instance is not safe for concurrent use;
// callers must serialize access through the embedded mutex.
type Runtime struct {
	mu       sync.Mutex
	runtime  wazero.Runtime
	module   api.Module
	compiled wazero.CompiledModule
}

// DecodeResult holds the result of a barcode decode operation.
type DecodeResult struct {
	Success      bool   `json:"success"`
	Text         string `json:"text"`
	Format       string `json:"format"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// EncodeResult holds the result of a barcode encode operation.
type EncodeResult struct {
	Success      bool    `json:"success"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Data         []uint8 `json:"data"`
	ErrorCode    int     `json:"error_code"`
	ErrorMessage string  `json:"error_message"`
}

// NewRuntime creates a new wazero WASM runtime instance.
func NewRuntime() *Runtime {
	return &Runtime{}
}

// Initialize loads and instantiates the WASM module from the given file path.
func (r *Runtime) Initialize(ctx context.Context, wasmPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.module != nil && !r.module.IsClosed() {
		return nil
	}

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file %s: %w", wasmPath, err)
	}

	config := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
	r.runtime = wazero.NewRuntimeWithConfig(ctx, config)

	// Instantiate WASI host functions (required by Emscripten STANDALONE_WASM)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r.runtime); err != nil {
		r.runtime.Close(context.Background())
		r.runtime = nil
		return fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Compile the WASM binary
	r.compiled, err = r.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		r.runtime.Close(context.Background())
		r.runtime = nil
		return fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Instantiate the WASM module
	// Disable _start to prevent proc_exit from being called automatically
	r.module, err = r.runtime.InstantiateModule(ctx, r.compiled, wazero.NewModuleConfig().WithStartFunctions())
	if err != nil {
		r.compiled.Close(context.Background())
		r.runtime.Close(context.Background())
		r.runtime = nil
		r.compiled = nil
		return fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	return nil
}

// IsReady returns whether the runtime has been initialized and the module is still usable.
func (r *Runtime) IsReady() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.module != nil && !r.module.IsClosed()
}

// DecodeImage decodes raw pixel data using the WASM module's decode_barcode_pixels export.
// The data parameter must be tightly packed pixel data with the given width, height, and channels.
// channels must be 1 (grayscale), 2 (grayscale+alpha), 3 (RGB), or 4 (RGBA).
func (r *Runtime) DecodeImage(ctx context.Context, data []byte, width, height, channels int, opts *DecodeOptions) (*DecodeResult, error) {
	if err := validateDecodeInput(data, width, height, channels); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.module == nil || r.module.IsClosed() {
		return nil, fmt.Errorf("WASM runtime not initialized or closed")
	}

	mem := r.module.Memory()
	if mem == nil {
		return nil, fmt.Errorf("WASM module has no exported memory")
	}

	// Allocate guest memory for pixel data
	pixelSize := width * height * channels
	pixelPtr, err := r.guestMalloc(ctx, uint64(pixelSize))
	if err != nil {
		return nil, err
	}
	defer r.guestFree(ctx, pixelPtr)

	// Write pixel data into guest memory
	if !mem.Write(pixelPtr, data[:pixelSize]) {
		return nil, fmt.Errorf("failed to write pixel data to WASM memory")
	}

	// Allocate and configure decode options
	optsPtr, err := r.guestCreateOptions(ctx)
	if err != nil {
		return nil, err
	}
	defer r.guestFreeOptions(ctx, optsPtr)

	if err := r.guestConfigureOptions(ctx, optsPtr, opts); err != nil {
		return nil, err
	}

	// Call decode_barcode_pixels(data, width, height, channels, options)
	decodeFn := r.module.ExportedFunction("decode_barcode_pixels")
	if decodeFn == nil {
		return nil, fmt.Errorf("decode_barcode_pixels not exported in WASM module")
	}

	resultRes, err := decodeFn.Call(ctx, uint64(pixelPtr), uint64(width), uint64(height), uint64(channels), uint64(optsPtr))
	if err != nil {
		// Cancellation may close the module; mark it unusable
		if ctx.Err() != nil {
			r.module = nil
		}
		return nil, fmt.Errorf("WASM decode_barcode_pixels call failed: %w", err)
	}
	resultPtr := uint32(resultRes[0])

	if resultPtr == 0 {
		errMsg := r.guestLastError(ctx, mem)
		return nil, fmt.Errorf("WASM decode failed: %s", errMsg)
	}
	defer r.guestFreeResult(ctx, resultPtr)

	// Read DecodeResult struct from WASM memory (wasm32, 12 bytes)
	resultBytes, ok := mem.Read(resultPtr, 12)
	if !ok {
		return nil, fmt.Errorf("failed to read result struct from WASM memory")
	}

	textPtr := binary.LittleEndian.Uint32(resultBytes[0:4])
	formatVal := int32(binary.LittleEndian.Uint32(resultBytes[4:8]))

	// Read text string (null-terminated, bounded read)
	text := ""
	if textPtr != 0 {
		textBytes, ok := mem.Read(textPtr, 4096)
		if ok {
			text = cString(textBytes)
		}
	}

	return &DecodeResult{
		Success: true,
		Text:    text,
		Format:  formatToString(int(formatVal)),
	}, nil
}

// validateDecodeInput checks pixel data, dimensions, and channels before entering WASM.
func validateDecodeInput(data []byte, width, height, channels int) error {
	if len(data) == 0 {
		return fmt.Errorf("empty image data")
	}
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}
	if channels < 1 || channels > 4 {
		return fmt.Errorf("unsupported channel count: %d (must be 1-4)", channels)
	}
	// Check for integer overflow: width * height * channels
	if width > (1<<30)/height/channels {
		return fmt.Errorf("image dimensions overflow: %dx%dx%d", width, height, channels)
	}
	expected := width * height * channels
	if len(data) < expected {
		return fmt.Errorf("data too short: have %d bytes, need %d", len(data), expected)
	}
	return nil
}

// guestMalloc allocates memory in the WASM guest via the zxing_malloc export.
func (r *Runtime) guestMalloc(ctx context.Context, size uint64) (uint32, error) {
	fn := r.module.ExportedFunction("zxing_malloc")
	if fn == nil {
		return 0, fmt.Errorf("zxing_malloc not exported in WASM module")
	}
	res, err := fn.Call(ctx, size)
	if err != nil {
		return 0, fmt.Errorf("failed to allocate WASM memory: %w", err)
	}
	ptr := uint32(res[0])
	if ptr == 0 {
		return 0, fmt.Errorf("WASM malloc returned null for size %d", size)
	}
	return ptr, nil
}

// guestFree releases guest memory allocated by guestMalloc.
func (r *Runtime) guestFree(ctx context.Context, ptr uint32) {
	if fn := r.module.ExportedFunction("zxing_free"); fn != nil {
		fn.Call(ctx, uint64(ptr))
	}
}

// guestCreateOptions allocates a default DecodeOptions struct in guest memory.
func (r *Runtime) guestCreateOptions(ctx context.Context) (uint32, error) {
	fn := r.module.ExportedFunction("create_default_options")
	if fn == nil {
		return 0, fmt.Errorf("create_default_options not exported in WASM module")
	}
	res, err := fn.Call(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create decode options: %w", err)
	}
	ptr := uint32(res[0])
	if ptr == 0 {
		return 0, fmt.Errorf("WASM create_default_options returned null")
	}
	return ptr, nil
}

// guestFreeOptions releases a DecodeOptions struct in guest memory.
func (r *Runtime) guestFreeOptions(ctx context.Context, ptr uint32) {
	if fn := r.module.ExportedFunction("free_options"); fn != nil {
		fn.Call(ctx, uint64(ptr))
	}
}

// guestConfigureOptions writes all DecodeOptions fields via the configure_decode_options export.
func (r *Runtime) guestConfigureOptions(ctx context.Context, optsPtr uint32, opts *DecodeOptions) error {
	fn := r.module.ExportedFunction("configure_decode_options")
	if fn == nil {
		return fmt.Errorf("configure_decode_options not exported in WASM module")
	}
	formats := 0xFFFF // FORMAT_ALL
	tryHarder := 1
	tryRotate := 1
	tryInvert := 0
	tryDownscale := 1
	if opts != nil {
		if opts.Formats != 0 {
			formats = opts.Formats
		}
		if opts.TryHarder {
			tryHarder = 1
		} else {
			tryHarder = 0
		}
		if opts.TryRotate {
			tryRotate = 1
		} else {
			tryRotate = 0
		}
		if opts.TryInvert {
			tryInvert = 1
		} else {
			tryInvert = 0
		}
		if opts.TryDownscale {
			tryDownscale = 1
		} else {
			tryDownscale = 0
		}
	}
	_, err := fn.Call(ctx, uint64(optsPtr), uint64(formats), uint64(tryHarder), uint64(tryRotate), uint64(tryInvert), uint64(tryDownscale))
	if err != nil {
		return fmt.Errorf("failed to configure decode options: %w", err)
	}
	return nil
}

// guestFreeResult releases a DecodeResult struct in guest memory.
func (r *Runtime) guestFreeResult(ctx context.Context, ptr uint32) {
	if fn := r.module.ExportedFunction("free_result"); fn != nil {
		fn.Call(ctx, uint64(ptr))
	}
}

// guestLastError reads the last error string from the WASM module.
func (r *Runtime) guestLastError(ctx context.Context, mem api.Memory) string {
	fn := r.module.ExportedFunction("get_last_error")
	if fn == nil {
		return "unknown error (get_last_error not exported)"
	}
	res, err := fn.Call(ctx)
	if err != nil || res == nil || res[0] == 0 {
		return "unknown error"
	}
	errBytes, ok := mem.Read(uint32(res[0]), 256)
	if !ok {
		return "failed to read error message"
	}
	return cString(errBytes)
}

// cString reads a null-terminated C string from a byte slice.
func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// formatToString converts a BarcodeFormat enum value to its string name.
func formatToString(format int) string {
	switch format {
	case 0:
		return "None"
	case 1:
		return "QR_CODE"
	case 2:
		return "AZTEC"
	case 4:
		return "CODABAR"
	case 8:
		return "CODE_39"
	case 16:
		return "CODE_93"
	case 32:
		return "CODE_128"
	case 64:
		return "DATA_MATRIX"
	case 128:
		return "EAN_8"
	case 256:
		return "EAN_13"
	case 512:
		return "ITF"
	case 1024:
		return "MAXICODE"
	case 2048:
		return "PDF_417"
	case 4096:
		return "UPC_A"
	case 8192:
		return "UPC_E"
	case 0xFFFF:
		return "ALL"
	default:
		return fmt.Sprintf("Unknown(%d)", format)
	}
}

// EncodeText encodes text to a barcode image using the WASM module.
// WASM encoding is not implemented yet; this always returns an error.
func (r *Runtime) EncodeText(text string, width, height int) (*EncodeResult, error) {
	return nil, fmt.Errorf("WASM encode is not implemented")
}

// Close releases all WASM runtime resources. It is idempotent.
func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	if r.compiled != nil {
		if err := r.compiled.Close(context.Background()); err != nil && firstErr == nil {
			firstErr = err
		}
		r.compiled = nil
	}
	if r.runtime != nil {
		if err := r.runtime.Close(context.Background()); err != nil && firstErr == nil {
			firstErr = err
		}
		r.runtime = nil
	}
	r.module = nil
	return firstErr
}
