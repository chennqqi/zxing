//go:build !cgo || !(linux || windows)

package wasm

import (
	"context"
	"image"
	"os"
	"testing"
)

func TestWazeroLoadAndDecode(t *testing.T) {
	rt := NewRuntime()
	err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm")
	if err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	if !rt.IsReady() {
		t.Fatal("Runtime should be ready after Initialize")
	}
}

func TestWazeroDecodeQRCode(t *testing.T) {
	rt := NewRuntime()
	err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm")
	if err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	// Load a real QR code test image
	file, err := os.Open("../../data/qrcode_www.bing.com.png")
	if err != nil {
		t.Fatalf("Failed to open test image: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode test image: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert to RGBA
	rgba := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	result, err := rt.DecodeImage(rgba.Pix, width, height, 4)
	if err != nil {
		t.Fatalf("Failed to decode QR code: %v", err)
	}

	if !result.Success {
		t.Fatalf("Decode failed: %s", result.ErrorMessage)
	}

	if len(result.Text) == 0 {
		t.Fatal("Decoded text should not be empty")
	}

	t.Logf("Decoded text: %s, format: %s", result.Text, result.Format)
}
