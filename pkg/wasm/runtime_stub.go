//go:build cgo && (linux || windows) && !(js && wasm)

package wasm

import (
	"context"
	"fmt"
)

// Runtime WASM 运行时（非 WASM 环境的存根实现）
type Runtime struct{}

// DecodeResult 解码结果
type DecodeResult struct {
	Text   string
	Format string
}

// EncodeResult 编码结果
type EncodeResult struct {
	Data   []byte
	Width  int
	Height int
}

// NewRuntime 创建新的 WASM 运行时
func NewRuntime() *Runtime {
	return &Runtime{}
}

// Initialize 初始化 WASM 运行时
func (r *Runtime) Initialize(ctx context.Context, wasmPath string) error {
	return fmt.Errorf("WASM runtime not available in non-WASM environment")
}

// DecodeImage 解码图像
func (r *Runtime) DecodeImage(data []byte, width, height, channels int) (*DecodeResult, error) {
	return nil, fmt.Errorf("WASM runtime not available in non-WASM environment")
}

// EncodeText 编码文本
func (r *Runtime) EncodeText(text string, width, height int) (*EncodeResult, error) {
	return nil, fmt.Errorf("WASM runtime not available in non-WASM environment")
}

// Close 关闭运行时
func (r *Runtime) Close() error {
	return nil
}
