//go:build !cgo || !(linux || windows || darwin)

// Package zxing CGO stub implementation for platforms where CGO is not available.
// This file provides stub types and functions so that the package compiles
// without CGO or on non-linux/windows/darwin platforms.
package zxing

import (
	"context"
	"fmt"
	"image"
)

// cgoAvailable indicates that CGO is not enabled on this platform.
// Used by factory.go for compile-time backend selection.
const cgoAvailable = false

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

// String returns the canonical UPPER_SNAKE name of the barcode format,
// consistent with Result.Format values on both CGO and WASM backends.
func (f BarcodeFormat) String() string {
	name := barcodeFormatName(f)
	if len(name) == 0 || name == "Unknown" {
		return fmt.Sprintf("Unknown(%d)", f)
	}
	return name
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
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// DecodeMulti returns an error when CGO is not available.
func DecodeMulti(imagePath string, options *CGODecodeOptions) ([]*CGODecodeResult, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
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
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// encodeWithCGOImpl provides a stub implementation when CGO is disabled.
func encodeWithCGOImpl(ctx context.Context, config *Config, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// cgoZXing is a stub that returns errors when CGO is not available.
type cgoZXing struct {
	config *Config
}

// DecodeImage returns an error when CGO is not available.
func (c *cgoZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// DecodeBytes returns an error when CGO is not available.
func (c *cgoZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// DecodeImageMulti returns an error when CGO is not available.
func (c *cgoZXing) DecodeImageMulti(ctx context.Context, img image.Image, opts *DecodeOptions) ([]*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// DecodeBytesMulti returns an error when CGO is not available.
func (c *cgoZXing) DecodeBytesMulti(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) ([]*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// EncodeText returns an error when CGO is not available.
func (c *cgoZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// EncodeToBytes returns an error when CGO is not available.
func (c *cgoZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	return nil, 0, 0, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin)")
}

// Close is a no-op stub.
func (c *cgoZXing) Close() error {
	return nil
}

// GetBackend returns the CGO backend type.
func (c *cgoZXing) GetBackend() Backend {
	return BackendCGO
}
