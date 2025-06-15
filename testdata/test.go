package testdata

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/threatbook/zxing"
)

// 生成测试用的二维码图片
func generateTestQRCode(t *testing.T, text string) string {
	// 创建临时目录
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")

	// 创建测试图片
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	defer file.Close()

	// 保存为 PNG
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}

	return filePath
}

func TestDecodeQRCode(t *testing.T) {
	// 生成测试图片
	imagePath := generateTestQRCode(t, "test")

	// 创建解码选项
	options := zxing.NewDefaultOptions()
	if options == nil {
		t.Fatal("Failed to create default options")
	}

	// 设置只识别二维码
	options.Formats = zxing.FormatQRCode

	// 解码
	result, err := zxing.Decode(imagePath, options)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// 验证结果
	if result == nil {
		t.Fatal("Decode returned nil result")
	}
	if result.Text == "" {
		t.Error("Decoded text is empty")
	}
	if result.Format != zxing.FormatQRCode {
		t.Errorf("Expected format QR Code, got %v", result.Format)
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Confidence out of range: %f", result.Confidence)
	}
}

func TestDecodeMultipleBarcodes(t *testing.T) {
	// 生成测试图片
	imagePath := generateTestQRCode(t, "test")

	// 创建解码选项
	options := zxing.NewDefaultOptions()
	if options == nil {
		t.Fatal("Failed to create default options")
	}

	// 设置支持所有格式
	options.Formats = zxing.FormatAll

	// 解码多个条码
	results, err := zxing.DecodeMulti(imagePath, options)
	if err != nil {
		t.Fatalf("Failed to decode multiple: %v", err)
	}

	// 验证结果
	if results == nil {
		t.Fatal("DecodeMulti returned nil results")
	}
	if len(results) == 0 {
		t.Error("No barcodes found")
	}

	// 检查每个结果
	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d is nil", i)
			continue
		}
		if result.Text == "" {
			t.Errorf("Result %d text is empty", i)
		}
		if result.Confidence < 0 || result.Confidence > 1 {
			t.Errorf("Result %d confidence out of range: %f", i, result.Confidence)
		}
	}
}

func TestDecodeOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  *zxing.DecodeOptions
		expected bool
	}{
		{
			name: "Default options",
			options: &zxing.DecodeOptions{
				Formats:      zxing.FormatAll,
				TryHarder:    true,
				TryRotate:    true,
				TryInvert:    false,
				TryDownscale: true,
			},
			expected: true,
		},
		{
			name: "QR Code only",
			options: &zxing.DecodeOptions{
				Formats:      zxing.FormatQRCode,
				TryHarder:    true,
				TryRotate:    true,
				TryInvert:    false,
				TryDownscale: true,
			},
			expected: true,
		},
		{
			name: "Try harder disabled",
			options: &zxing.DecodeOptions{
				Formats:      zxing.FormatAll,
				TryHarder:    false,
				TryRotate:    true,
				TryInvert:    false,
				TryDownscale: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成测试图片
			imagePath := generateTestQRCode(t, "test")

			// 解码
			result, err := zxing.Decode(imagePath, tt.options)
			if err != nil {
				if tt.expected {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if !tt.expected {
				t.Error("Expected error, got nil")
				return
			}

			// 验证结果
			if result == nil {
				t.Error("Decode returned nil result")
				return
			}
			if result.Text == "" {
				t.Error("Decoded text is empty")
			}
			if result.Confidence < 0 || result.Confidence > 1 {
				t.Errorf("Confidence out of range: %f", result.Confidence)
			}
		})
	}
}
