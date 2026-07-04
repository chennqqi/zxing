//go:build cgo && (linux || windows) && !(js && wasm)

// Package zxing WASM stub for CGO-enabled platforms.
// When CGO is enabled on linux/windows, the WASM backend is not available.
package zxing

import (
	"context"
	"fmt"
	"image"
)

// wasmZXing is a stub that returns errors when WASM backend is not available.
type wasmZXing struct {
	config *Config
}

// DecodeImage returns an error on CGO platforms.
func (w *wasmZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("WASM backend is not available when CGO is enabled on linux/windows")
}

// DecodeBytes returns an error on CGO platforms.
func (w *wasmZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("WASM backend is not available when CGO is enabled on linux/windows")
}

// EncodeText returns an error on CGO platforms.
func (w *wasmZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("WASM backend is not available when CGO is enabled on linux/windows")
}

// EncodeToBytes returns an error on CGO platforms.
func (w *wasmZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	return nil, 0, 0, fmt.Errorf("WASM backend is not available when CGO is enabled on linux/windows")
}

// Close is a no-op stub.
func (w *wasmZXing) Close() error {
	return nil
}

// GetBackend returns the WASM backend type.
func (w *wasmZXing) GetBackend() Backend {
	return BackendWASM
}
