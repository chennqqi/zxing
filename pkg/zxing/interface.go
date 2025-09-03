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
}

// Decoder 解码器接口
type Decoder interface {
	// DecodeImage 解码图像
	DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error)
	
	// DecodeBytes 解码字节数据
	DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error)
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