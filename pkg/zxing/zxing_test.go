package zxing

import (
	"context"
	"image"
	_ "image/png"
	"os"
	"testing"
	"time"
)

func TestNewZXing(t *testing.T) {
	config := DefaultConfig()
	config.Backend = BackendAuto

	zx, err := New(config)
	if err != nil {
		t.Fatalf("failed to create ZXing instance: %v", err)
	}
	defer zx.Close()

	// Backend should be either CGO or WASM depending on build tags
	backend := zx.GetBackend()
	if backend != BackendCGO && backend != BackendWASM {
		t.Errorf("expected backend CGO or WASM, got %s", backend)
	}
	t.Logf("Selected backend: %s", backend)
}

func TestBackendSelection(t *testing.T) {
	tests := []struct {
		name     string
		backend  Backend
		expected Backend
	}{
		{"CGO backend", BackendCGO, BackendCGO},
		{"WASM backend", BackendWASM, BackendWASM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Backend = tt.backend

			zx, err := New(config)
			if err != nil {
				t.Fatalf("failed to create instance: %v", err)
			}
			defer zx.Close()

			if zx.GetBackend() != tt.expected {
				t.Errorf("expected backend %s, got %s", tt.expected, zx.GetBackend())
			}
		})
	}
}

func TestDecodeQRCodeImage(t *testing.T) {
	config := DefaultConfig()
	config.Backend = BackendAuto

	zx, err := New(config)
	if err != nil {
		t.Fatalf("failed to create ZXing instance: %v", err)
	}
	defer zx.Close()

	// Load a real QR code test image
	file, err := os.Open("../../data/qrcode_www.bing.com.png")
	if err != nil {
		t.Skipf("test image not found: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("failed to decode test image: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := zx.DecodeImage(ctx, img, &DecodeOptions{TryHarder: true})
	if err != nil {
		t.Fatalf("failed to decode QR code: %v", err)
	}

	if len(result.Text) == 0 {
		t.Error("decoded text should not be empty")
	}

	t.Logf("Decoded text: %s, format: %s, backend: %s", result.Text, result.Format, zx.GetBackend())
}

func TestConfigFromEnv(t *testing.T) {
	config := DefaultConfig()

	if config.Backend != BackendAuto {
		t.Errorf("expected default backend %s, got %s", BackendAuto, config.Backend)
	}

	if config.Timeout != 30 {
		t.Errorf("expected default timeout 30, got %d", config.Timeout)
	}

	if config.Debug != false {
		t.Errorf("expected default debug false, got %t", config.Debug)
	}

	if config.WASMPath != "wasm/zxingwrapper.wasm" {
		t.Errorf("expected default WASM path wasm/zxingwrapper.wasm, got %s", config.WASMPath)
	}
}
