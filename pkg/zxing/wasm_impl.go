//go:build (!cgo || !(linux || windows)) && !(js && wasm)

// Package zxing WASM backend implementation using wazero runtime.
// This file is active when CGO is disabled or on non-linux/windows platforms.
package zxing

import (
	"context"
	"fmt"
	"image"
	"image/color"

	"github.com/chennqqi/zxing/pkg/wasm"
)

// wasmZXing implements the ZXing interface using the wazero WASM runtime.
type wasmZXing struct {
	config  *Config
	runtime *wasm.Runtime
}

// ensureRuntime lazily initializes the WASM runtime.
func (w *wasmZXing) ensureRuntime(ctx context.Context) error {
	if w.runtime == nil {
		w.runtime = wasm.NewRuntime()
		if err := w.runtime.Initialize(ctx, w.config.WASMPath); err != nil {
			return fmt.Errorf("failed to initialize WASM runtime: %w", err)
		}
	}
	return nil
}

// DecodeImage decodes an image using the WASM backend.
func (w *wasmZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	if err := w.ensureRuntime(ctx); err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert image to RGBA byte data
	data := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			idx := (y*width + x) * 4
			data[idx] = uint8(r >> 8)
			data[idx+1] = uint8(g >> 8)
			data[idx+2] = uint8(b >> 8)
			data[idx+3] = uint8(a >> 8)
		}
	}

	return w.DecodeBytes(ctx, data, width, height, opts)
}

// DecodeBytes decodes raw RGBA byte data using the WASM backend.
func (w *wasmZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	if err := w.ensureRuntime(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	result, err := w.runtime.DecodeImage(data, width, height, 4)
	if err != nil {
		return nil, fmt.Errorf("WASM decode failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("decode failed: %s (code: %d)", result.ErrorMessage, result.ErrorCode)
	}

	return &Result{
		Text:   result.Text,
		Format: result.Format,
		Points: []image.Point{},
		Metadata: map[string]interface{}{
			"backend": "wasm",
		},
	}, nil
}

// EncodeText encodes text to a barcode image using the WASM backend.
func (w *wasmZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	if err := w.ensureRuntime(ctx); err != nil {
		return nil, err
	}

	if len(text) == 0 {
		return nil, fmt.Errorf("empty text")
	}

	if opts == nil {
		opts = &EncodeOptions{
			Width:  256,
			Height: 256,
			Format: "QR_CODE",
		}
	}

	result, err := w.runtime.EncodeText(text, opts.Width, opts.Height)
	if err != nil {
		return nil, fmt.Errorf("WASM encode failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("encode failed: %s (code: %d)", result.ErrorMessage, result.ErrorCode)
	}

	// Convert byte data to image
	img := image.NewGray(image.Rect(0, 0, result.Width, result.Height))
	for i, val := range result.Data {
		if i >= len(result.Data) {
			break
		}
		x := i % result.Width
		y := i / result.Width
		if x < result.Width && y < result.Height {
			img.SetGray(x, y, color.Gray{Y: val})
		}
	}

	return img, nil
}

// EncodeToBytes encodes text to raw byte data using the WASM backend.
func (w *wasmZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	if err := w.ensureRuntime(ctx); err != nil {
		return nil, 0, 0, err
	}

	if len(text) == 0 {
		return nil, 0, 0, fmt.Errorf("empty text")
	}

	if opts == nil {
		opts = &EncodeOptions{
			Width:  256,
			Height: 256,
			Format: "QR_CODE",
		}
	}

	result, err := w.runtime.EncodeText(text, opts.Width, opts.Height)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("WASM encode failed: %w", err)
	}

	if !result.Success {
		return nil, 0, 0, fmt.Errorf("encode failed: %s (code: %d)", result.ErrorMessage, result.ErrorCode)
	}

	return result.Data, result.Width, result.Height, nil
}

// Close releases WASM runtime resources.
func (w *wasmZXing) Close() error {
	if w.runtime != nil {
		return w.runtime.Close()
	}
	return nil
}

// GetBackend returns the backend type.
func (w *wasmZXing) GetBackend() Backend {
	return BackendWASM
}
