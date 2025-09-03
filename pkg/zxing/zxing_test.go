package zxing

import (
	"context"
	"testing"
	"time"
)

func TestNewZXing(t *testing.T) {
	config := DefaultConfig()
	config.Backend = BackendCGO

	zx, err := New(config)
	if err != nil {
		t.Fatalf("创建 ZXing 实例失败: %v", err)
	}
	defer zx.Close()

	if zx.GetBackend() != BackendCGO {
		t.Errorf("期望后端 %s，实际 %s", BackendCGO, zx.GetBackend())
	}
}

func TestEncodeText(t *testing.T) {
	config := DefaultConfig()
	config.Backend = BackendCGO

	zx, err := New(config)
	if err != nil {
		t.Fatalf("创建 ZXing 实例失败: %v", err)
	}
	defer zx.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := &EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	img, err := zx.EncodeText(ctx, "Hello, ZXing!", opts)
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 256 {
		t.Errorf("期望图像尺寸 256x256，实际 %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestDecodeBytes(t *testing.T) {
	config := DefaultConfig()
	config.Backend = BackendCGO

	zx, err := New(config)
	if err != nil {
		t.Fatalf("创建 ZXing 实例失败: %v", err)
	}
	defer zx.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建测试数据
	width, height := 256, 256
	data := make([]byte, width*height*4) // RGBA

	opts := &DecodeOptions{
		TryHarder: true,
	}

	result, err := zx.DecodeBytes(ctx, data, width, height, opts)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	if len(result.Text) == 0 {
		t.Error("解码结果为空")
	}

	if len(result.Format) == 0 {
		t.Error("格式信息为空")
	}
}

func TestBackendSelection(t *testing.T) {
	tests := []struct {
		name     string
		backend  Backend
		expected Backend
	}{
		{"CGO后端", BackendCGO, BackendCGO},
		{"WASM后端", BackendWASM, BackendWASM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Backend = tt.backend

			zx, err := New(config)
			if err != nil {
				t.Fatalf("创建实例失败: %v", err)
			}
			defer zx.Close()

			if zx.GetBackend() != tt.expected {
				t.Errorf("期望后端 %s，实际 %s", tt.expected, zx.GetBackend())
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	// 测试默认配置
	config := DefaultConfig()

	if config.Backend != BackendAuto {
		t.Errorf("期望默认后端 %s，实际 %s", BackendAuto, config.Backend)
	}

	if config.Timeout != 30 {
		t.Errorf("期望默认超时 30，实际 %d", config.Timeout)
	}

	if config.Debug != false {
		t.Errorf("期望默认调试模式 false，实际 %t", config.Debug)
	}
}

func BenchmarkEncode(b *testing.B) {
	config := DefaultConfig()
	config.Backend = BackendCGO

	zx, err := New(config)
	if err != nil {
		b.Fatalf("创建实例失败: %v", err)
	}
	defer zx.Close()

	ctx := context.Background()
	opts := &EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := zx.EncodeText(ctx, "Benchmark test", opts)
		if err != nil {
			b.Fatalf("编码失败: %v", err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	config := DefaultConfig()
	config.Backend = BackendCGO

	zx, err := New(config)
	if err != nil {
		b.Fatalf("创建实例失败: %v", err)
	}
	defer zx.Close()

	ctx := context.Background()

	// 准备测试数据
	width, height := 256, 256
	data := make([]byte, width*height*4)
	opts := &DecodeOptions{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := zx.DecodeBytes(ctx, data, width, height, opts)
		if err != nil {
			b.Fatalf("解码失败: %v", err)
		}
	}
}
