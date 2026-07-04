//go:build js && wasm

// WASM 实现
package zxing

import (
	"context"
	"fmt"
	"image"
	"image/color"

	"github.com/chennqqi/zxing/pkg/wasm"
)

// wasmZXing WASM 实现
type wasmZXing struct {
	config  *Config
	runtime *wasm.Runtime
}

// DecodeImage 解码图像
func (w *wasmZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
	if !w.runtime.IsReady() {
		return nil, fmt.Errorf("WASM runtime not ready")
	}

	// 转换图像为字节数据
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	data := make([]byte, width*height*4) // RGBA

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

	return w.DecodeBytes(ctx, data, width, height, opts)
}

// DecodeBytes 解码字节数据
func (w *wasmZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	if !w.runtime.IsReady() {
		return nil, fmt.Errorf("WASM runtime not ready")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	// 调用 WASM 解码函数
	result, err := w.runtime.DecodeImage(data, width, height, 4)
	if err != nil {
		return nil, fmt.Errorf("WASM decode failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("decode failed: %s (code: %d)", result.ErrorMessage, result.ErrorCode)
	}

	return &Result{
		Text:     result.Text,
		Format:   result.Format,
		Points:   []image.Point{}, // TODO: 从 WASM 结果中提取位置信息
		Metadata: make(map[string]interface{}),
	}, nil
}

// EncodeText 编码文本为条码图像
func (w *wasmZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	if !w.runtime.IsReady() {
		return nil, fmt.Errorf("WASM runtime not ready")
	}

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

	// 调用 WASM 编码函数
	result, err := w.runtime.EncodeText(text, opts.Width, opts.Height)
	if err != nil {
		return nil, fmt.Errorf("WASM encode failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("encode failed: %s (code: %d)", result.ErrorMessage, result.ErrorCode)
	}

	// 转换字节数据为图像
	img := image.NewGray(image.Rect(0, 0, result.Width, result.Height))

	for i, val := range result.Data {
		if i >= len(result.Data) {
			break
		}

		x := i % result.Width
		y := i / result.Width

		if x < result.Width && y < result.Height {
			img.SetGray(x, y, color.Gray{Y: val})
		}
	}

	return img, nil
}

// EncodeToBytes 编码文本为字节数据
func (w *wasmZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	if !w.runtime.IsReady() {
		return nil, 0, 0, fmt.Errorf("WASM runtime not ready")
	}

	if len(text) == 0 {
		return nil, 0, 0, fmt.Errorf("empty text")
	}

	if opts == nil {
		opts = &EncodeOptions{
			Width:  256,
			Height: 256,
			Format: "QR_CODE",
		}
	}

	// 调用 WASM 编码函数
	result, err := w.runtime.EncodeText(text, opts.Width, opts.Height)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("WASM encode failed: %w", err)
	}

	if !result.Success {
		return nil, 0, 0, fmt.Errorf("encode failed: %s (code: %d)", result.ErrorMessage, result.ErrorCode)
	}

	return result.Data, result.Width, result.Height, nil
}

// Close 关闭资源
func (w *wasmZXing) Close() error {
	if w.runtime != nil {
		return w.runtime.Close()
	}
	return nil
}

// GetBackend 获取当前使用的后端类型
func (w *wasmZXing) GetBackend() Backend {
	return BackendWASM
}
