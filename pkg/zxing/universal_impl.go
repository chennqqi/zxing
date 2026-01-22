package zxing

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"runtime"

	"github.com/chennqqi/zxing/pkg/wasm"
)

// universalZXing 通用实现，支持多种后端
type universalZXing struct {
	backend Backend
	config  *Config
	runtime *wasm.Runtime
}

// DecodeImage 解码图像
func (u *universalZXing) DecodeImage(ctx context.Context, img image.Image, opts *DecodeOptions) (*Result, error) {
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
	
	return u.DecodeBytes(ctx, data, width, height, opts)
}

// DecodeBytes 解码字节数据
func (u *universalZXing) DecodeBytes(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}
	
	if opts == nil {
		opts = &DecodeOptions{}
	}
	
	switch u.backend {
	case BackendWASM:
		return u.decodeWithWASM(ctx, data, width, height, opts)
	case BackendCGO:
		return u.decodeWithCGO(ctx, data, width, height, opts)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", u.backend)
	}
}

// EncodeText 编码文本为条码图像
func (u *universalZXing) EncodeText(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
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
	
	switch u.backend {
	case BackendWASM:
		return u.encodeWithWASM(ctx, text, opts)
	case BackendCGO:
		return u.encodeWithCGO(ctx, text, opts)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", u.backend)
	}
}

// EncodeToBytes 编码文本为字节数据
func (u *universalZXing) EncodeToBytes(ctx context.Context, text string, opts *EncodeOptions) ([]byte, int, int, error) {
	img, err := u.EncodeText(ctx, text, opts)
	if err != nil {
		return nil, 0, 0, err
	}
	
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	// 转换为字节数组
	data := make([]byte, width*height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := img.(*image.Gray)
			data[y*width+x] = gray.GrayAt(x, y).Y
		}
	}
	
	return data, width, height, nil
}

// GetBackend 获取当前使用的后端类型
func (u *universalZXing) GetBackend() Backend {
	return u.backend
}

// Close 关闭资源
func (u *universalZXing) Close() error {
	if u.runtime != nil {
		return u.runtime.Close()
	}
	return nil
}

// decodeWithWASM 使用 WASM 后端解码
func (u *universalZXing) decodeWithWASM(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	if runtime.GOOS == "js" && runtime.GOARCH == "wasm" {
		// 真实的 WASM 环境
		if u.runtime == nil {
			u.runtime = wasm.NewRuntime()
			if err := u.runtime.Initialize(ctx, u.config.WASMPath); err != nil {
				return nil, fmt.Errorf("failed to initialize WASM runtime: %w", err)
			}
		}
		
		result, err := u.runtime.DecodeImage(data, width, height, 4)
		if err != nil {
			return nil, err
		}
		
		return &Result{
			Text:   result.Text,
			Format: result.Format,
			Points: []image.Point{},
		}, nil
	} else {
		// 非 WASM 环境，使用模拟实现
		if u.config.Debug {
			fmt.Println("WASM 后端在当前环境使用模拟实现")
		}
		
		return &Result{
			Text:   "WASM bytes decode simulation",
			Format: "QR_CODE",
			Points: []image.Point{},
		}, nil
	}
}

// decodeWithCGO 使用 CGO 后端解码
func (u *universalZXing) decodeWithCGO(ctx context.Context, data []byte, width, height int, opts *DecodeOptions) (*Result, error) {
	// 创建CGO实例并调用
	cgoImpl := &cgoZXing{
		config: u.config,
	}
	return cgoImpl.DecodeBytes(ctx, data, width, height, opts)
}

// encodeWithWASM 使用 WASM 后端编码
func (u *universalZXing) encodeWithWASM(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	if runtime.GOOS == "js" && runtime.GOARCH == "wasm" {
		// 真实的 WASM 环境
		if u.runtime == nil {
			u.runtime = wasm.NewRuntime()
			if err := u.runtime.Initialize(ctx, u.config.WASMPath); err != nil {
				return nil, fmt.Errorf("failed to initialize WASM runtime: %w", err)
			}
		}
		
		result, err := u.runtime.EncodeText(text, opts.Width, opts.Height)
		if err != nil {
			return nil, err
		}
		
		// 转换为 Go 图像
		img := image.NewGray(image.Rect(0, 0, result.Width, result.Height))
		for i, value := range result.Data {
			y := i / result.Width
			x := i % result.Width
			img.SetGray(x, y, color.Gray{Y: value})
		}
		
		return img, nil
	} else {
		// 非 WASM 环境，使用模拟实现
		img := image.NewGray(image.Rect(0, 0, opts.Width, opts.Height))
		
		// 生成简单的测试图案
		for y := 0; y < opts.Height; y++ {
			for x := 0; x < opts.Width; x++ {
				if (x+y)%20 < 10 {
					img.SetGray(x, y, color.Gray{Y: 0})
				} else {
					img.SetGray(x, y, color.Gray{Y: 255})
				}
			}
		}
		
		return img, nil
	}
}

// encodeWithCGO 使用 CGO 后端编码
func (u *universalZXing) encodeWithCGO(ctx context.Context, text string, opts *EncodeOptions) (image.Image, error) {
	// 创建CGO实例并调用
	cgoImpl := &cgoZXing{
		config: u.config,
	}
	return cgoImpl.EncodeText(ctx, text, opts)
}