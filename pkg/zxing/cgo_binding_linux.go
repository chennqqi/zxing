//go:build cgo && linux

// Package zxing CGO binding for Linux platform.
// This file provides the Cgo declarations and type definitions specific to Linux.
package zxing

/*
#cgo CXXFLAGS: -std=c++20
#cgo CFLAGS: -I${SRCDIR}/../../include
#cgo LDFLAGS: -L${SRCDIR}/../../lib/linux-x64 -lzxingwrapper -lZXing -lstdc++ -lm
#include <stdlib.h>
#include "zxing.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// BarcodeFormat represents the barcode format type.
type BarcodeFormat int

const (
	FormatNone       BarcodeFormat = C.FORMAT_NONE
	FormatQRCode     BarcodeFormat = C.FORMAT_QR_CODE
	FormatAztec      BarcodeFormat = C.FORMAT_AZTEC
	FormatCodabar    BarcodeFormat = C.FORMAT_CODABAR
	FormatCode39     BarcodeFormat = C.FORMAT_CODE_39
	FormatCode93     BarcodeFormat = C.FORMAT_CODE_93
	FormatCode128    BarcodeFormat = C.FORMAT_CODE_128
	FormatDataMatrix BarcodeFormat = C.FORMAT_DATA_MATRIX
	FormatEAN8       BarcodeFormat = C.FORMAT_EAN_8
	FormatEAN13      BarcodeFormat = C.FORMAT_EAN_13
	FormatITF        BarcodeFormat = C.FORMAT_ITF
	FormatMaxiCode   BarcodeFormat = C.FORMAT_MAXICODE
	FormatPDF417     BarcodeFormat = C.FORMAT_PDF_417
	FormatUPCA       BarcodeFormat = C.FORMAT_UPC_A
	FormatUPCE       BarcodeFormat = C.FORMAT_UPC_E
	FormatAll        BarcodeFormat = C.FORMAT_ALL
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

// CGODecodeOptions represents CGO decode options.
type CGODecodeOptions struct {
	Formats      BarcodeFormat
	TryHarder    bool
	TryRotate    bool
	TryInvert    bool
	TryDownscale bool
}

// CGODecodeResult represents a CGO decode result.
type CGODecodeResult struct {
	Text       string
	Format     BarcodeFormat
	Confidence float32
}

// NewDefaultOptions creates default decode options via CGO.
func NewDefaultOptions() *CGODecodeOptions {
	options := C.create_default_options()
	if options == nil {
		return nil
	}
	defer C.free_options(options)

	return &CGODecodeOptions{
		Formats:      BarcodeFormat(options.formats),
		TryHarder:    options.try_harder != 0,
		TryRotate:    options.try_rotate != 0,
		TryInvert:    options.try_invert != 0,
		TryDownscale: options.try_downscale != 0,
	}
}

// Decode decodes a single barcode from the given image path.
func Decode(imagePath string, options *CGODecodeOptions) (*CGODecodeResult, error) {
	if options == nil {
		options = NewDefaultOptions()
		if options == nil {
			return nil, fmt.Errorf("failed to create default options")
		}
	}

	cPath := C.CString(imagePath)
	defer C.free(unsafe.Pointer(cPath))

	cOptions := C.DecodeOptions{
		formats:       C.int(options.Formats),
		try_harder:    C.int(boolToInt(options.TryHarder)),
		try_rotate:    C.int(boolToInt(options.TryRotate)),
		try_invert:    C.int(boolToInt(options.TryInvert)),
		try_downscale: C.int(boolToInt(options.TryDownscale)),
	}

	result := C.decode_barcode(cPath, &cOptions)
	if result == nil {
		return nil, fmt.Errorf("%s", C.GoString(C.get_last_error()))
	}
	defer C.free_result(result)

	return &CGODecodeResult{
		Text:       C.GoString(result.text),
		Format:     BarcodeFormat(result.format),
		Confidence: float32(result.confidence),
	}, nil
}

// DecodeMulti decodes multiple barcodes from the given image path.
func DecodeMulti(imagePath string, options *CGODecodeOptions) ([]*CGODecodeResult, error) {
	if options == nil {
		options = NewDefaultOptions()
		if options == nil {
			return nil, fmt.Errorf("failed to create default options")
		}
	}

	cPath := C.CString(imagePath)
	defer C.free(unsafe.Pointer(cPath))

	cOptions := C.DecodeOptions{
		formats:       C.int(options.Formats),
		try_harder:    C.int(boolToInt(options.TryHarder)),
		try_rotate:    C.int(boolToInt(options.TryRotate)),
		try_invert:    C.int(boolToInt(options.TryInvert)),
		try_downscale: C.int(boolToInt(options.TryDownscale)),
	}

	var count C.int
	results := C.decode_barcodes(cPath, &cOptions, &count)
	if results == nil {
		return nil, fmt.Errorf("%s", C.GoString(C.get_last_error()))
	}
	defer C.free_results(results, count)

	goResults := make([]*CGODecodeResult, int(count))
	for i := 0; i < int(count); i++ {
		result := C.decode_result_get(results, C.int(i))
		if result == nil {
			continue
		}

		goResults[i] = &CGODecodeResult{
			Text:       C.GoString(result.text),
			Format:     BarcodeFormat(result.format),
			Confidence: float32(result.confidence),
		}
	}

	return goResults, nil
}

// boolToInt converts a bool to an int (1 or 0).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
