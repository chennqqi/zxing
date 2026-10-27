// Package zxing 定义统一的接口
package zxing

import (
	"context"
	"image"
)

// Result 解码结果
type Result struct {
	// Text 解码得到的文本内容
	Text string

	// Format 条码格式（如 QR_CODE, CODE_128 等）
	Format string

	// Points 条码在图像中的位置点
	Points []image.Point

	// Metadata 额外的元数据信息
	Metadata map[string]interface{}
}

// EncodeOptions 编码选项
type EncodeOptions struct {
	// Width 生成图像的宽度
	Width int

	// Height 生成图像的高度
	Height int

	// Format 条码格式
	Format string

	// ErrorCorrectionLevel 错误纠正级别（适用于二维码）
	ErrorCorrectionLevel string

	// Margin 边距大小
	Margin int
}

// DecodeOptions 解码选项
type DecodeOptions struct {
	// TryHarder 是否尝试更努力地解码
	TryHarder bool

	// PossibleFormats 可能的格式列表
	PossibleFormats []string

	// CharacterSet 字符集
	CharacterSet string

	// MaxResults 多码解码时返回的最大结果数量，<= 0 表示不限制
	MaxResults int
}

// Decoder 解码器接口
type Decoder interface {
	// DecodeImage 解码图像
	DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error)

	// DecodeBytes 解码字节数据
	DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error)

	// DecodeImageMulti 解码图像中的所有条码，未识别到码时返回空集合
	DecodeImageMulti(ctx context.Context, img image.Image, opts *DecodeOptions) ([]*Result, error)

	// DecodeBytesMulti 解码字节数据中的所有条码，未识别到码时返回空集合
	DecodeBytesMulti(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) ([]*Result, error)
}

// Encoder 编码器接口
type Encoder interface {
	// EncodeText 编码文本为条码图像
	EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error)

	// EncodeToBytes 编码文本为字节数据
	EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error)
}

// ZXing 主接口，组合了编码器和解码器
type ZXing interface {
	Decoder
	Encoder

	// Close 关闭资源
	Close() error

	// GetBackend 获取当前使用的后端类型
	GetBackend() Backend
}

// barcodeFormatFlag maps a format name to its C API flag bit.
// It accepts both UPPER_SNAKE ("QR_CODE") and ZXing-style ("QRCode") spellings.
func barcodeFormatFlag(name string) BarcodeFormat {
	switch name {
	case "QR_CODE", "QRCode":
		return FormatQRCode
	case "AZTEC", "Aztec":
		return FormatAztec
	case "CODABAR", "Codabar":
		return FormatCodabar
	case "CODE_39", "Code39":
		return FormatCode39
	case "CODE_93", "Code93":
		return FormatCode93
	case "CODE_128", "Code128":
		return FormatCode128
	case "DATA_MATRIX", "DataMatrix":
		return FormatDataMatrix
	case "EAN_8", "EAN8":
		return FormatEAN8
	case "EAN_13", "EAN13":
		return FormatEAN13
	case "ITF":
		return FormatITF
	case "MAXICODE", "MaxiCode":
		return FormatMaxiCode
	case "PDF_417", "PDF417":
		return FormatPDF417
	case "UPC_A", "UPCA":
		return FormatUPCA
	case "UPC_E", "UPCE":
		return FormatUPCE
	}
	return FormatNone
}

// barcodeFormatName returns the canonical UPPER_SNAKE name of a format flag,
// consistent with the format names returned by the WASM backend.
func barcodeFormatName(f BarcodeFormat) string {
	switch f {
	case FormatNone:
		return "None"
	case FormatQRCode:
		return "QR_CODE"
	case FormatAztec:
		return "AZTEC"
	case FormatCodabar:
		return "CODABAR"
	case FormatCode39:
		return "CODE_39"
	case FormatCode93:
		return "CODE_93"
	case FormatCode128:
		return "CODE_128"
	case FormatDataMatrix:
		return "DATA_MATRIX"
	case FormatEAN8:
		return "EAN_8"
	case FormatEAN13:
		return "EAN_13"
	case FormatITF:
		return "ITF"
	case FormatMaxiCode:
		return "MAXICODE"
	case FormatPDF417:
		return "PDF_417"
	case FormatUPCA:
		return "UPC_A"
	case FormatUPCE:
		return "UPC_E"
	case FormatAll:
		return "ALL"
	}
	return "Unknown"
}
