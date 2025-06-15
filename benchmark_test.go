// Package zxing 提供了条码解码功能
package zxing

import (
	"testing"
)

const (
	FormatNone       = 0
	FormatQRCode     = 1
	FormatAztec      = 2
	FormatCodabar    = 4
	FormatCode39     = 8
	FormatCode93     = 16
	FormatCode128    = 32
	FormatDataMatrix = 64
	FormatEAN8       = 128
	FormatEAN13      = 256
	FormatITF        = 512
	FormatMaxiCode   = 1024
	FormatPDF417     = 2048
	FormatUPCA       = 4096
	FormatUPCE       = 8192
	FormatAll        = 0xFFFF
)

// BenchmarkDecode 测试单个条码解码性能
// 使用默认选项解码单个二维码
func BenchmarkDecode(b *testing.B) {
	opts := NewDefaultOptions()
	if opts == nil {
		b.Fatal("Failed to create default options")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := Decode("testdata/qrcode.png", opts)
		if err != nil {
			b.Fatal(err)
		}
		if result == nil {
			b.Fatal("Decode result is nil")
		}
	}
}

// BenchmarkDecodeMulti 测试多个条码解码性能
// 使用默认选项解码包含多个条码的图片
func BenchmarkDecodeMulti(b *testing.B) {
	opts := NewDefaultOptions()
	if opts == nil {
		b.Fatal("Failed to create default options")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := DecodeMulti("testdata/multi.png", opts)
		if err != nil {
			b.Fatal(err)
		}
		if results == nil {
			b.Fatal("Decode results is nil")
		}
	}
}

// BenchmarkDecodeWithOptions 测试带高级选项的单条码解码性能
// 使用自定义选项（只解码QR码、启用更努力的解码模式、启用图像旋转）
func BenchmarkDecodeWithOptions(b *testing.B) {
	opts := NewDefaultOptions()
	if opts == nil {
		b.Fatal("Failed to create default options")
	}

	// 设置高级选项
	opts.Formats = FormatQRCode
	opts.TryHarder = true
	opts.TryRotate = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := Decode("testdata/qrcode.png", opts)
		if err != nil {
			b.Fatal(err)
		}
		if result == nil {
			b.Fatal("Decode result is nil")
		}
	}
}

// BenchmarkDecodeMultiWithOptions 测试带高级选项的多条码解码性能
// 使用自定义选项（解码所有格式、启用更努力的解码模式、启用图像旋转）
func BenchmarkDecodeMultiWithOptions(b *testing.B) {
	opts := NewDefaultOptions()
	if opts == nil {
		b.Fatal("Failed to create default options")
	}

	// 设置高级选项
	opts.Formats = FormatAll
	opts.TryHarder = true
	opts.TryRotate = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := DecodeMulti("testdata/multi.png", opts)
		if err != nil {
			b.Fatal(err)
		}
		if results == nil {
			b.Fatal("Decode results is nil")
		}
	}
}
