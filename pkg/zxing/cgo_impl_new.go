//go:build cgo

// CGO 实现 - 集成现有的CGO代码
package zxing

/*
#cgo CXXFLAGS: -std=c++17 -I. -I./include -I./zxing-cpp/core/src
#cgo LDFLAGS: -L./lib/windows/x64 -L./lib/linux/x64 -lzxingwrapper -lZXing -lstdc++
#include <stdlib.h>
#include "include/zxing.h"
*/
import "C"
import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"unsafe"
)

// cgoZXing CGO 实现
type cgoZXing struct {
	config *Config
}

// DecodeImage 解码图像
func (c *cgoZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	if opts == nil {
		opts = &DecodeOptions{}
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 转换图像为字节数据
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

// DecodeBytes 解码字节数据
func (c *cgoZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	if opts == nil {
		opts = &DecodeOptions{}
	}

	// 创建临时文件来保存图像数据
	tempFile, err := os.CreateTemp("", "zxing_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 将字节数据编码为PNG图像并写入临时文件
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			if idx+3 < len(data) {
				img.Set(x, y, color.RGBA{
					R: data[idx],
					G: data[idx+1],
					B: data[idx+2],
					A: data[idx+3],
				})
			}
		}
	}
	
	// 编码为PNG
	encoder := &png.Encoder{}
	if err := encoder.Encode(tempFile, img); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	tempFile.Close()

	// 调用CGO解码函数
	cPath := C.CString(tempFile.Name())
	defer C.free(unsafe.Pointer(cPath))

	// 创建C解码选项
	cOptions := C.DecodeOptions{
		formats:       C.int(C.FORMAT_ALL),
		try_harder:    C.int(boolToInt(opts.TryHarder)),
		try_rotate:    C.int(1), // 默认启用旋转
		try_invert:    C.int(0), // 默认不启用反转
		try_downscale: C.int(1), // 默认启用缩放
	}

	// 调用C函数解码
	result := C.decode_barcode(cPath, &cOptions)
	if result == nil {
		errorMsg := C.GoString(C.get_last_error())
		return nil, fmt.Errorf("CGO decode failed: %s", errorMsg)
	}
	defer C.free_result(result)

	// 转换结果格式
	return &Result{
		Text:   C.GoString(result.text),
		Format: formatToString(C.int(result.format)),
		Points: []image.Point{}, // C API暂不支持位置信息
		Metadata: map[string]interface{}{
			"confidence": float32(result.confidence),
			"backend":    "cgo",
		},
	}, nil
}

// EncodeText 编码文本为条码图像
func (c *cgoZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	if len(text) == 0 {
		return nil, fmt.Errorf("empty text")
	}

	if opts == nil {
		opts = &EncodeOptions{
			Width:  256,
			Height: 256,
			Format: "QR_CODE",
		}
	}

	// CGO编码功能暂未实现，因为zxing-cpp主要专注于解码
	// 如果需要编码功能，建议使用WASM后端或其他专门的编码库
	// 这里返回一个占位图像
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// 填充白色背景
	for y := 0; y < opts.Height; y++ {
		for x := 0; x < opts.Width; x++ {
			img.Set(x, y, color.White)
		}
	}

	return img, fmt.Errorf("encoding not implemented in CGO backend, please use WASM backend for encoding")
}

// EncodeToBytes 编码文本为字节数据
func (c *cgoZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	img, err := c.EncodeText(ctx, text, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 转换为字节数组
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

// Close 关闭资源
func (c *cgoZXing) Close() error {
	// CGO 实现的清理逻辑
	return nil
}

// GetBackend 获取当前使用的后端类型
func (c *cgoZXing) GetBackend() Backend {
	return BackendCGO
}

// 辅助函数：将bool转换为int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// 辅助函数：将C格式枚举转换为字符串
func formatToString(format C.int) string {
	switch format {
	case C.FORMAT_QR_CODE:
		return "QR_CODE"
	case C.FORMAT_CODE_128:
		return "CODE_128"
	case C.FORMAT_CODE_39:
		return "CODE_39"
	case C.FORMAT_CODE_93:
		return "CODE_93"
	case C.FORMAT_EAN_8:
		return "EAN_8"
	case C.FORMAT_EAN_13:
		return "EAN_13"
	case C.FORMAT_UPC_A:
		return "UPC_A"
	case C.FORMAT_UPC_E:
		return "UPC_E"
	case C.FORMAT_DATA_MATRIX:
		return "DATA_MATRIX"
	case C.FORMAT_PDF_417:
		return "PDF_417"
	case C.FORMAT_AZTEC:
		return "AZTEC"
	case C.FORMAT_CODABAR:
		return "CODABAR"
	case C.FORMAT_ITF:
		return "ITF"
	case C.FORMAT_MAXICODE:
		return "MAXICODE"
	default:
		return "UNKNOWN"
	}
}
