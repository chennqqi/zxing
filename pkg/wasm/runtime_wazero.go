//go:build (!cgo || !(linux || windows)) && !(js && wasm)

// Package wasm provides ZXing WebAssembly runtime support via wazero.
package wasm

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runtime manages the wazero WASM runtime for ZXing.
type Runtime struct {
	runtime wazero.Runtime
	module  api.Module
	ready   bool
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
	if r.ready {
		return nil
	}

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file %s: %w", wasmPath, err)
	}

	r.runtime = wazero.NewRuntime(ctx)

	// Instantiate WASI host functions (required by Emscripten STANDALONE_WASM)
	wasi_snapshot_preview1.MustInstantiate(ctx, r.runtime)

	// Compile the WASM binary
	compiled, err := r.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Instantiate the WASM module
	// Disable _start to prevent proc_exit from being called automatically
	r.module, err = r.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithStartFunctions())
	if err != nil {
		return fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	r.ready = true
	return nil
}

// IsReady returns whether the runtime has been initialized.
func (r *Runtime) IsReady() bool {
	return r.ready
}

// DecodeImage decodes image data (RGBA format) using the WASM module.
// The data parameter should be RGBA pixel data with the given width and height.
func (r *Runtime) DecodeImage(data []byte, width, height, channels int) (*DecodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// Encode RGBA data to PNG in Go (WASM's stb_image can decode PNG from memory)
	pngData, err := encodeRGBAtoPNG(data, width, height, channels)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	ctx := context.Background()
	allocFn := r.module.ExportedFunction("zxing_alloc")

	if allocFn == nil {
		return nil, fmt.Errorf("zxing_alloc not exported in WASM module")
	}

	// Reset bump allocator before use
	if resetFn := r.module.ExportedFunction("zxing_alloc_reset"); resetFn != nil {
		resetFn.Call(ctx)
	}

	// Allocate memory for PNG data in WASM
	ptrRes, err := allocFn.Call(ctx, uint64(len(pngData)))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate WASM memory: %w", err)
	}
	pngPtr := uint32(ptrRes[0])

	// Write PNG data to WASM memory
	mem := r.module.Memory()
	if !mem.Write(pngPtr, pngData) {
		return nil, fmt.Errorf("failed to write PNG data to WASM memory")
	}

	// Create default decode options
	optsRes, err := r.module.ExportedFunction("create_default_options").Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create options: %w", err)
	}
	optsPtr := optsRes[0]
	if optsPtr == 0 {
		return nil, fmt.Errorf("failed to allocate decode options")
	}
	defer r.module.ExportedFunction("free_options").Call(ctx, optsPtr)

	// Call decode_barcode_data(pngPtr, pngLen, optsPtr)
	decodeFn := r.module.ExportedFunction("decode_barcode_data")
	if decodeFn == nil {
		return nil, fmt.Errorf("decode_barcode_data not exported in WASM module")
	}

	resultRes, err := decodeFn.Call(ctx, uint64(pngPtr), uint64(len(pngData)), optsPtr)
	if err != nil {
		return nil, fmt.Errorf("WASM decode_barcode_data call failed: %w", err)
	}
	resultPtr := uint32(resultRes[0])

	if resultPtr == 0 {
		// Get error message from get_last_error
		errFn := r.module.ExportedFunction("get_last_error")
		if errFn != nil {
			errRes, _ := errFn.Call(ctx)
			if errRes != nil && errRes[0] != 0 {
				errBytes, ok := mem.Read(uint32(errRes[0]), 256)
				if ok {
					return nil, fmt.Errorf("WASM decode error: %s", cString(errBytes))
				}
			}
		}
		return nil, fmt.Errorf("WASM decode failed with no error message")
	}
	defer r.module.ExportedFunction("free_result").Call(ctx, uint64(resultPtr))

	// Read DecodeResult struct from WASM memory
	// C struct layout (wasm32, 4-byte aligned):
	//   char* text;        // 4 bytes (pointer)
	//   BarcodeFormat format; // 4 bytes (enum/int)
	//   float confidence;  // 4 bytes
	// Total: 12 bytes
	resultBytes, ok := mem.Read(resultPtr, 12)
	if !ok {
		return nil, fmt.Errorf("failed to read result struct from WASM memory")
	}

	textPtr := binary.LittleEndian.Uint32(resultBytes[0:4])
	formatVal := int32(binary.LittleEndian.Uint32(resultBytes[4:8]))

	// Read text string (null-terminated)
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

// encodeRGBAtoPNG encodes raw pixel data to PNG format.
func encodeRGBAtoPNG(data []byte, width, height, channels int) ([]byte, error) {
	if channels == 4 {
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		copy(img.Pix, data)
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	} else if channels == 3 {
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				idx := (y*width + x) * 3
				pixIdx := (y*width + x) * 4
				if idx+2 < len(data) {
					img.Pix[pixIdx] = data[idx]
					img.Pix[pixIdx+1] = data[idx+1]
					img.Pix[pixIdx+2] = data[idx+2]
					img.Pix[pixIdx+3] = 255
				}
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, fmt.Errorf("unsupported channel count: %d", channels)
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
// This is a PoC stub — full implementation in later tasks.
func (r *Runtime) EncodeText(text string, width, height int) (*EncodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}
	return &EncodeResult{
		Success:      false,
		ErrorCode:    -1,
		ErrorMessage: "PoC: encode not fully implemented yet",
	}, nil
}

// Close releases all WASM runtime resources.
func (r *Runtime) Close() error {
	if r.runtime != nil {
		ctx := context.Background()
		r.runtime.Close(ctx)
		r.runtime = nil
	}
	r.module = nil
	r.ready = false
	return nil
}
