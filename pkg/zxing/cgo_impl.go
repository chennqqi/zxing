//go:build cgo && (linux || windows || darwin)

// Package zxing CGO implementation.
// This file contains the CGO-backed ZXing implementation that uses the
// platform-conditional binding file (cgo_binding.go)
// for Cgo declarations and type definitions.
package zxing

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
)

// cgoAvailable indicates that CGO is enabled on this platform.
// Used by factory.go for compile-time backend selection.
const cgoAvailable = true

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
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	if opts == nil {
		opts = &DecodeOptions{}
	}

	// Create a temporary PNG file for CGO decode (C API reads from file path)
	path, err := decodeTempPNG(data, width, height)
	if err != nil {
		return nil, err
	}
	defer os.Remove(path)

	cgoOpts, err := buildCGOOptions(opts)
	if err != nil {
		return nil, err
	}

	// Call CGO decode
	result, err := Decode(path, cgoOpts)
	if err != nil {
		return nil, err
	}

	return &Result{
		Text:   result.Text,
		Format: barcodeFormatName(result.Format),
		Points: []image.Point{},
		Metadata: map[string]interface{}{
			"confidence": result.Confidence,
			"backend":    "cgo",
		},
	}, nil
}

// buildCGOOptions converts public DecodeOptions to CGODecodeOptions.
func buildCGOOptions(opts *DecodeOptions) (*CGODecodeOptions, error) {
	cgoOpts := NewDefaultOptions()
	if cgoOpts == nil {
		return nil, fmt.Errorf("failed to create CGO options")
	}
	cgoOpts.TryHarder = opts.TryHarder
	formatFlags := FormatNone
	for _, format := range opts.PossibleFormats {
		formatFlags |= barcodeFormatFlag(format)
	}
	if formatFlags == FormatNone {
		cgoOpts.Formats = FormatAll
	} else {
		cgoOpts.Formats = formatFlags
	}
	return cgoOpts, nil
}

// decodeTempPNG writes raw RGBA data to a temporary PNG file and returns its path.
// The C decode_barcodes API reads images from file paths, so an intermediate
// PNG is required until the static libraries export a raw-pixel multi decoder.
func decodeTempPNG(data []byte, width, height int) (string, error) {
	tempFile, err := os.CreateTemp("", "zxing_*.png")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	if len(data) >= len(img.Pix) {
		copy(img.Pix, data)
	} else {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("image data size mismatch")
	}

	if err := png.Encode(tempFile, img); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}
	return tempFile.Name(), nil
}

// DecodeImageMulti decodes all barcodes in an image using the CGO backend.
func (c *cgoZXing) DecodeImageMulti(ctx context.Context, img image.Image, opts *DecodeOptions) ([]*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
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

	return c.DecodeBytesMulti(ctx, data, width, height, opts)
}

// DecodeBytesMulti decodes all barcodes in raw RGBA byte data using the CGO backend.
// It returns an empty (non-nil) slice when no barcode is found.
// Note: the file-path based C API does not return barcode corner positions,
// so Result.Points is empty for the CGO backend.
func (c *cgoZXing) DecodeBytesMulti(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) ([]*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	if opts == nil {
		opts = &DecodeOptions{}
	}

	path, err := decodeTempPNG(data, width, height)
	if err != nil {
		return nil, err
	}
	defer os.Remove(path)

	cgoOpts, err := buildCGOOptions(opts)
	if err != nil {
		return nil, err
	}

	cgoResults, err := DecodeMulti(path, cgoOpts)
	if err != nil {
		// The C API reports "no barcode found" through the last-error channel;
		// an absent result list means an empty set for the caller.
		return []*Result{}, nil
	}

	results := make([]*Result, 0, len(cgoResults))
	for _, r := range cgoResults {
		if r == nil {
			continue
		}
		results = append(results, &Result{
			Text:   r.Text,
			Format: barcodeFormatName(r.Format),
			Points: []image.Point{},
			Metadata: map[string]interface{}{
				"confidence": r.Confidence,
				"backend":    "cgo",
			},
		})
	}

	// Apply the MaxResults limit on the Go side; the C API has no such option.
	if opts.MaxResults > 0 && len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return results, nil
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
