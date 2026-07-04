//go:build !cgo || !(linux || windows)

// Package wasm provides ZXing WebAssembly runtime support via wazero.
package wasm

import (
	"context"
	"fmt"
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
// This is a PoC stub — full implementation in task 3.
func (r *Runtime) DecodeImage(data []byte, width, height, channels int) (*DecodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}
	return &DecodeResult{
		Success:      false,
		ErrorCode:    -1,
		ErrorMessage: "PoC: decode not fully implemented yet",
	}, nil
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
