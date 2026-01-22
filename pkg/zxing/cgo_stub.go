//go:build !cgo

// CGO stub implementation for when CGO is disabled
package zxing

import (
	"context"
	"fmt"
	"image"
)

// decodeWithCGOImpl provides a stub implementation when CGO is disabled
func decodeWithCGOImpl(ctx context.Context, config *Config, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 and cgo build tag)")
}

// encodeWithCGOImpl provides a stub implementation when CGO is disabled
func encodeWithCGOImpl(ctx context.Context, config *Config, text string, opts *EncodeOptions) (image.Image, error) {
	return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 and cgo build tag)")
}
