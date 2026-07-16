//go:build !cgo || !(linux || windows)

package wasm

import (
	"context"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"sync"
	"testing"
	"time"
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

func TestWazeroRequiredExports(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	for _, name := range []string{"zxing_malloc", "zxing_free", "configure_decode_options", "decode_barcode_pixels", "create_default_options", "free_options", "free_result", "get_last_error"} {
		if rt.module.ExportedFunction(name) == nil {
			t.Fatalf("required export %q is missing", name)
		}
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

	result, err := rt.DecodeImage(context.Background(), rgba.Pix, width, height, 4, nil)
	if err != nil {
		t.Fatalf("Failed to decode QR code: %v", err)
	}

	if !result.Success {
		t.Fatalf("Decode failed: %s", result.ErrorMessage)
	}

	if len(result.Text) == 0 {
		t.Fatal("Decoded text should not be empty")
	}
	if result.Format != "QR_CODE" {
		t.Fatalf("unexpected decoded format: %q", result.Format)
	}

	t.Logf("Decoded text: %s, format: %s", result.Text, result.Format)
}

func TestDecodeImageRejectsInvalidInput(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	tests := []struct {
		name     string
		data     []byte
		width    int
		height   int
		channels int
	}{
		{"empty data", nil, 10, 10, 4},
		{"zero width", make([]byte, 40), 0, 10, 4},
		{"zero height", make([]byte, 40), 10, 0, 4},
		{"channel 0", make([]byte, 100), 10, 10, 0},
		{"channel 5", make([]byte, 500), 10, 10, 5},
		{"undersized buffer", []byte{1, 2, 3}, 1, 1, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := rt.DecodeImage(context.Background(), tt.data, tt.width, tt.height, tt.channels, nil)
			if err == nil {
				t.Fatal("expected error for invalid input, got nil")
			}
		})
	}
}

func TestRuntimeConcurrentDecode(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

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

	rgba := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	const goroutines = 8
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			result, err := rt.DecodeImage(context.Background(), rgba.Pix, width, height, 4, nil)
			if err != nil {
				t.Errorf("decode failed: %v", err)
				return
			}
			if !result.Success {
				t.Errorf("decode unsuccessful: %s", result.ErrorMessage)
			}
			if result.Format != "QR_CODE" {
				t.Errorf("unexpected format: %s", result.Format)
			}
		}()
	}
	wg.Wait()
}

func TestRuntimeCloseIsConcurrentAndIdempotent(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func() {
			defer wg.Done()
			if err := rt.Close(); err != nil {
				t.Errorf("close returned error: %v", err)
			}
		}()
	}
	wg.Wait()

	if rt.IsReady() {
		t.Fatal("runtime should not be ready after close")
	}
}

func TestDecodeImageHonorsCanceledContext(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rt.DecodeImage(ctx, make([]byte, 16), 2, 2, 4, nil)
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
}

func TestInitializeFailureLeavesRuntimeNotReady(t *testing.T) {
	rt := NewRuntime()
	err := rt.Initialize(context.Background(), "/nonexistent/path/to/wasm.wasm")
	if err == nil {
		t.Fatal("expected error for nonexistent WASM path")
	}
	if rt.IsReady() {
		t.Fatal("runtime should not be ready after failed init")
	}
}

func TestDecodeWithOptions(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

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

	rgba := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	opts := &DecodeOptions{
		Formats:      1, // FORMAT_QR_CODE
		TryHarder:    true,
		TryRotate:    true,
		TryInvert:    false,
		TryDownscale: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := rt.DecodeImage(ctx, rgba.Pix, width, height, 4, opts)
	if err != nil {
		t.Fatalf("Failed to decode QR code with options: %v", err)
	}

	if !result.Success {
		t.Fatalf("Decode failed: %s", result.ErrorMessage)
	}

	if result.Format != "QR_CODE" {
		t.Fatalf("unexpected decoded format: %q", result.Format)
	}

	t.Logf("Decoded text: %s, format: %s", result.Text, result.Format)
}

func TestDecodeWithIncompatibleFormat(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

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

	rgba := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	opts := &DecodeOptions{
		Formats: 2, // FORMAT_AZTEC only — should not find QR code
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := rt.DecodeImage(ctx, rgba.Pix, width, height, 4, opts)
	if err != nil {
		// Error is expected when no barcode of the specified format is found
		t.Logf("Expected error with incompatible format: %v", err)
		return
	}
	if result.Success {
		t.Fatalf("expected decode to fail with incompatible format, got success: %s", result.Text)
	}
	t.Logf("Expected decode failure with incompatible format: %s", result.ErrorMessage)
}

func TestEncodeTextReturnsError(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	_, err := rt.EncodeText("test", 256, 256)
	if err == nil {
		t.Fatal("expected error from unimplemented EncodeText")
	}
	t.Logf("EncodeText correctly returned error: %v", err)
}

func TestCloseReturnsNoErrorWhenAlreadyClosed(t *testing.T) {
	rt := NewRuntime()
	if err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm"); err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}

	if err := rt.Close(); err != nil {
		t.Fatalf("first close returned error: %v", err)
	}
	if err := rt.Close(); err != nil {
		t.Fatalf("second close returned error: %v", err)
	}
}

func TestFormatToString(t *testing.T) {
	tests := []struct {
		format int
		want   string
	}{
		{0, "None"}, {1, "QR_CODE"}, {2, "AZTEC"}, {4, "CODABAR"},
		{8, "CODE_39"}, {16, "CODE_93"}, {32, "CODE_128"},
		{64, "DATA_MATRIX"}, {128, "EAN_8"}, {256, "EAN_13"},
		{512, "ITF"}, {1024, "MAXICODE"}, {2048, "PDF_417"},
		{4096, "UPC_A"}, {8192, "UPC_E"}, {0xFFFF, "ALL"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.format), func(t *testing.T) {
			got := formatToString(tt.format)
			if got != tt.want {
				t.Errorf("formatToString(%d) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}
