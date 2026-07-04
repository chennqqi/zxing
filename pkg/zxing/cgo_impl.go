//go:build cgo && (linux || windows)

// Package zxing CGO implementation.
// This file contains the CGO-backed ZXing implementation that uses the
// platform-specific binding files (cgo_binding_linux.go / cgo_binding_windows.go)
// for Cgo declarations and type definitions.
package zxing

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
)

// cgoZXing implements the ZXing interface using CGO.
type cgoZXing struct {
	config *Config
}

// DecodeImage decodes an image using the CGO backend.
func (c *cgoZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	if opts == nil {
		opts = &DecodeOptions{}
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

	return c.DecodeBytes(ctx, data, width, height, opts)
}

// DecodeBytes decodes raw RGBA byte data using the CGO backend.
func (c *cgoZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	if opts == nil {
		opts = &DecodeOptions{}
	}

	// Create a temporary PNG file for CGO decode (C API reads from file path)
	tempFile, err := os.CreateTemp("", "zxing_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Encode RGBA data to PNG
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	if len(data) >= len(img.Pix) {
		copy(img.Pix, data)
	} else {
		return nil, fmt.Errorf("image data size mismatch")
	}

	if err := png.Encode(tempFile, img); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	tempFile.Close()

	// Build CGO options
	cgoOpts := NewDefaultOptions()
	if cgoOpts == nil {
		return nil, fmt.Errorf("failed to create CGO options")
	}
	cgoOpts.TryHarder = opts.TryHarder
	if len(opts.PossibleFormats) > 0 {
		formatFlags := FormatNone
		for _, format := range opts.PossibleFormats {
			switch format {
			case "QR_CODE", "QRCode":
				formatFlags |= FormatQRCode
			case "AZTEC":
				formatFlags |= FormatAztec
			case "DATA_MATRIX", "DataMatrix":
				formatFlags |= FormatDataMatrix
			case "PDF_417", "PDF417":
				formatFlags |= FormatPDF417
			case "CODE_128", "Code128":
				formatFlags |= FormatCode128
			case "CODE_39", "Code39":
				formatFlags |= FormatCode39
			case "CODE_93", "Code93":
				formatFlags |= FormatCode93
			case "EAN_8", "EAN8":
				formatFlags |= FormatEAN8
			case "EAN_13", "EAN13":
				formatFlags |= FormatEAN13
			case "UPC_A", "UPCA":
				formatFlags |= FormatUPCA
			case "UPC_E", "UPCE":
				formatFlags |= FormatUPCE
			case "ITF":
				formatFlags |= FormatITF
			case "CODABAR":
				formatFlags |= FormatCodabar
			case "MAXICODE":
				formatFlags |= FormatMaxiCode
			}
		}
		if formatFlags == FormatNone {
			cgoOpts.Formats = FormatAll
		} else {
			cgoOpts.Formats = formatFlags
		}
	} else {
		cgoOpts.Formats = FormatAll
	}

	// Call CGO decode
	result, err := Decode(tempFile.Name(), cgoOpts)
	if err != nil {
		return nil, err
	}

	return &Result{
		Text:   result.Text,
		Format: result.Format.String(),
		Points: []image.Point{},
		Metadata: map[string]interface{}{
			"confidence": result.Confidence,
			"backend":    "cgo",
		},
	}, nil
}

// EncodeText encodes text to a barcode image using the CGO backend.
func (c *cgoZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("encoding is not supported by the CGO backend")
}

// EncodeToBytes encodes text to raw byte data using the CGO backend.
func (c *cgoZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	img, err := c.EncodeText(ctx, text, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

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

	return data, width, height, nil
}

// Close releases CGO backend resources.
func (c *cgoZXing) Close() error {
	return nil
}

// GetBackend returns the backend type.
func (c *cgoZXing) GetBackend() Backend {
	return BackendCGO
}

// decodeWithCGOImpl is called by universal_impl to delegate to CGO backend.
func decodeWithCGOImpl(ctx context.Context, config *Config, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	cgoImpl := &cgoZXing{
		config: config,
	}
	return cgoImpl.DecodeBytes(ctx, data, width, height, opts)
}

// encodeWithCGOImpl is called by universal_impl to delegate to CGO backend.
func encodeWithCGOImpl(ctx context.Context, config *Config, text string, opts *EncodeOptions) (image.Image, error) {
	cgoImpl := &cgoZXing{
		config: config,
	}
	return cgoImpl.EncodeText(ctx, text, opts)
}
