package zxing

import (
	"testing"
)

func TestNewDefaultOptions(t *testing.T) {
	options := NewDefaultOptions()
	if options == nil {
		t.Fatal("NewDefaultOptions returned nil")
	}

	// 检查默认值
	if options.Formats != FormatAll {
		t.Errorf("Expected Formats to be FormatAll, got %v", options.Formats)
	}
	if !options.TryHarder {
		t.Error("Expected TryHarder to be true")
	}
	if !options.TryRotate {
		t.Error("Expected TryRotate to be true")
	}
	if options.TryInvert {
		t.Error("Expected TryInvert to be false")
	}
	if !options.TryDownscale {
		t.Error("Expected TryDownscale to be true")
	}
}

func TestDecodeWithNilOptions(t *testing.T) {
	// 测试使用 nil 选项
	result, err := Decode("test.png", nil)
	if err == nil {
		t.Error("Expected error when options is nil")
	}
	if result != nil {
		t.Error("Expected nil result when options is nil")
	}
}

func TestDecodeWithInvalidPath(t *testing.T) {
	options := NewDefaultOptions()
	if options == nil {
		t.Fatal("Failed to create default options")
	}

	// 测试无效的文件路径
	result, err := Decode("nonexistent.png", options)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if result != nil {
		t.Error("Expected nil result for nonexistent file")
	}
}

func TestDecodeMultiWithNilOptions(t *testing.T) {
	// 测试使用 nil 选项
	results, err := DecodeMulti("test.png", nil)
	if err == nil {
		t.Error("Expected error when options is nil")
	}
	if results != nil {
		t.Error("Expected nil results when options is nil")
	}
}

func TestDecodeMultiWithInvalidPath(t *testing.T) {
	options := NewDefaultOptions()
	if options == nil {
		t.Fatal("Failed to create default options")
	}

	// 测试无效的文件路径
	results, err := DecodeMulti("nonexistent.png", options)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if results != nil {
		t.Error("Expected nil results for nonexistent file")
	}
}

func TestBarcodeFormatString(t *testing.T) {
	tests := []struct {
		format BarcodeFormat
		want   string
	}{
		{FormatNone, "None"},
		{FormatQRCode, "QR Code"},
		{FormatAztec, "Aztec"},
		{FormatCodabar, "Codabar"},
		{FormatCode39, "Code 39"},
		{FormatCode93, "Code 93"},
		{FormatCode128, "Code 128"},
		{FormatDataMatrix, "Data Matrix"},
		{FormatEAN8, "EAN-8"},
		{FormatEAN13, "EAN-13"},
		{FormatITF, "ITF"},
		{FormatMaxiCode, "MaxiCode"},
		{FormatPDF417, "PDF417"},
		{FormatUPCA, "UPC-A"},
		{FormatUPCE, "UPC-E"},
		{FormatAll, "All"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("BarcodeFormat.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
