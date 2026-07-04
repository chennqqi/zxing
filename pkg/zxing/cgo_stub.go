//go:build !cgo || !(linux || windows)

// Package zxing CGO stub implementation for platforms where CGO is not available.
// This file provides stub types and functions so that the package compiles
// without CGO or on non-linux/windows platforms.
package zxing

import (
	"context"
	"fmt"
	"image"
)

// BarcodeFormat represents the barcode format type (stub values for non-CGO builds).
type BarcodeFormat int

const (
	FormatNone       BarcodeFormat = 0
	FormatQRCode     BarcodeFormat = 1
	FormatAztec      BarcodeFormat = 2
	FormatCodabar    BarcodeFormat = 4
	FormatCode39     BarcodeFormat = 8
	FormatCode93     BarcodeFormat = 16
	FormatCode128    BarcodeFormat = 32
	FormatDataMatrix BarcodeFormat = 64
	FormatEAN8       BarcodeFormat = 128
	FormatEAN13      BarcodeFormat = 256
	FormatITF        BarcodeFormat = 512
	FormatMaxiCode   BarcodeFormat = 1024
	FormatPDF417     BarcodeFormat = 2048
	FormatUPCA       BarcodeFormat = 4096
	FormatUPCE       BarcodeFormat = 8192
	FormatAll        BarcodeFormat = 0xFFFF
)

// String returns the string representation of the barcode format.
func (f BarcodeFormat) String() string {
	switch f {
	case FormatNone:
		return "None"
	case FormatQRCode:
		return "QR Code"
	case FormatAztec:
		return "Aztec"
	case FormatCodabar:
		return "Codabar"
	case FormatCode39:
		return "Code 39"
	case FormatCode93:
		return "Code 93"
	case FormatCode128:
		return "Code 128"
	case FormatDataMatrix:
		return "Data Matrix"
	case FormatEAN8:
		return "EAN-8"
	case FormatEAN13:
		return "EAN-13"
	case FormatITF:
		return "ITF"
	case FormatMaxiCode:
		return "MaxiCode"
	case FormatPDF417:
		return "PDF417"
	case FormatUPCA:
		return "UPC-A"
	case FormatUPCE:
		return "UPC-E"
	case FormatAll:
		return "All"
	default:
		return fmt.Sprintf("Unknown(%d)", f)
	}
}

// CGODecodeOptions represents CGO decode options (stub for non-CGO builds).
type CGODecodeOptions struct {
	Formats      BarcodeFormat
	TryHarder    bool
	TryRotate    bool
	TryInvert    bool
	TryDownscale bool
}

// CGODecodeResult represents a CGO decode result (stub for non-CGO builds).
type CGODecodeResult struct {
	Text       string
	Format     BarcodeFormat
	Confidence float32
}

// NewDefaultOptions returns nil when CGO is not available.
func NewDefaultOptions() *CGODecodeOptions {
	return nil
}

// Decode returns an error when CGO is not available.
func Decode(imagePath string, options *CGODecodeOptions) (*CGODecodeResult, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// DecodeMulti returns an error when CGO is not available.
func DecodeMulti(imagePath string, options *CGODecodeOptions) ([]*CGODecodeResult, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// boolToInt converts a bool to an int (1 or 0).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// decodeWithCGOImpl provides a stub implementation when CGO is disabled.
func decodeWithCGOImpl(ctx context.Context, config *Config, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// encodeWithCGOImpl provides a stub implementation when CGO is disabled.
func encodeWithCGOImpl(ctx context.Context, config *Config, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// cgoZXing is a stub that returns errors when CGO is not available.
type cgoZXing struct {
	config *Config
}

// DecodeImage returns an error when CGO is not available.
func (c *cgoZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// DecodeBytes returns an error when CGO is not available.
func (c *cgoZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// EncodeText returns an error when CGO is not available.
func (c *cgoZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// EncodeToBytes returns an error when CGO is not available.
func (c *cgoZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	return nil, 0, 0, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux or windows)")
}

// Close is a no-op stub.
func (c *cgoZXing) Close() error {
	return nil
}

// GetBackend returns the CGO backend type.
func (c *cgoZXing) GetBackend() Backend {
	return BackendCGO
}
