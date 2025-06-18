package main

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L. -lzxing
// #include <stdlib.h>
// #include "include/zxing.h"
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// BarcodeFormat 表示条码格式
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

// String 返回条码格式的字符串表示
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

// DecodeOptions 表示解码选项
type DecodeOptions struct {
	Formats      BarcodeFormat
	TryHarder    bool
	TryRotate    bool
	TryInvert    bool
	TryDownscale bool
}

// DecodeResult 表示解码结果
type DecodeResult struct {
	Text       string
	Format     BarcodeFormat
	Confidence float32
}

// NewDefaultOptions 创建默认解码选项
func NewDefaultOptions() *DecodeOptions {
	options := C.create_default_options()
	if options == nil {
		return nil
	}
	defer C.free_options(options)

	return &DecodeOptions{
		Formats:      BarcodeFormat(options.formats),
		TryHarder:    options.try_harder != 0,
		TryRotate:    options.try_rotate != 0,
		TryInvert:    options.try_invert != 0,
		TryDownscale: options.try_downscale != 0,
	}
}

// Decode 解码单个条码
func Decode(imagePath string, options *DecodeOptions) (*DecodeResult, error) {
	if options == nil {
		options = NewDefaultOptions()
		if options == nil {
			return nil, errors.New("failed to create default options")
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
		return nil, errors.New(C.GoString(C.get_last_error()))
	}
	defer C.free_result(result)

	return &DecodeResult{
		Text:       C.GoString(result.text),
		Format:     BarcodeFormat(result.format),
		Confidence: float32(result.confidence),
	}, nil
}

// DecodeMulti 解码多个条码
func DecodeMulti(imagePath string, options *DecodeOptions) ([]*DecodeResult, error) {
	if options == nil {
		options = NewDefaultOptions()
		if options == nil {
			return nil, errors.New("failed to create default options")
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
		return nil, errors.New(C.GoString(C.get_last_error()))
	}
	defer C.free_results(results, count)

	goResults := make([]*DecodeResult, int(count))
	for i := 0; i < int(count); i++ {
		result := C.decode_result_get(results, C.int(i))
		if result == nil {
			continue
		}

		goResults[i] = &DecodeResult{
			Text:       C.GoString(result.text),
			Format:     BarcodeFormat(result.format),
			Confidence: float32(result.confidence),
		}
	}

	return goResults, nil
}

// 辅助函数：将 bool 转换为 int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// main 函数，用于满足 c-shared 构建要求
func main() {
	// 空的 main 函数，仅用于满足 buildmode=c-shared 的要求
}
