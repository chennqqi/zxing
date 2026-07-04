//go:build js && wasm
// +build js,wasm

// Package wasm 提供 ZXing 的 WebAssembly 运行时支持
package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
)

// Runtime WASM 运行时管理器
type Runtime struct {
	module js.Value
	ready  bool
}

// DecodeResult 解码结果
type DecodeResult struct {
	Success      bool   `json:"success"`
	Text         string `json:"text"`
	Format       string `json:"format"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// EncodeResult 编码结果
type EncodeResult struct {
	Success      bool    `json:"success"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Data         []uint8 `json:"data"`
	ErrorCode    int     `json:"error_code"`
	ErrorMessage string  `json:"error_message"`
}

// NewRuntime 创建新的 WASM 运行时
func NewRuntime() *Runtime {
	return &Runtime{
		ready: false,
	}
}

// Initialize 初始化 WASM 模块
func (r *Runtime) Initialize(ctx context.Context, wasmPath string) error {
	if r.ready {
		return nil
	}

	// 检查 WebAssembly 支持
	if !js.Global().Get("WebAssembly").Truthy() {
		return fmt.Errorf("WebAssembly not supported in this environment")
	}

	// 加载 WASM 模块
	promise := js.Global().Call("fetch", wasmPath)

	// 等待加载完成
	done := make(chan error, 1)

	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		if !response.Get("ok").Bool() {
			done <- fmt.Errorf("failed to fetch WASM module: %s", response.Get("statusText").String())
			return nil
		}

		arrayBufferPromise := response.Call("arrayBuffer")
		arrayBufferPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			arrayBuffer := args[0]

			instantiatePromise := js.Global().Get("WebAssembly").Call("instantiate", arrayBuffer)
			instantiatePromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				result := args[0]
				r.module = result.Get("instance")
				r.ready = true
				done <- nil
				return nil
			})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				done <- fmt.Errorf("failed to instantiate WASM: %v", args[0])
				return nil
			}))

			return nil
		})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			done <- fmt.Errorf("failed to get array buffer: %v", args[0])
			return nil
		}))

		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		done <- fmt.Errorf("failed to fetch: %v", args[0])
		return nil
	}))

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsReady 检查运行时是否就绪
func (r *Runtime) IsReady() bool {
	return r.ready
}

// DecodeImage 解码图像数据
func (r *Runtime) DecodeImage(imageData []byte, width, height, channels int) (*DecodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// 创建 JavaScript 数组
	jsArray := js.Global().Get("Uint8Array").New(len(imageData))
	js.CopyBytesToJS(jsArray, imageData)

	// 调用 WASM 函数
	result := r.module.Get("exports").Call("decode_image_data", jsArray, width, height, channels)

	// 解析结果
	resultJSON := js.Global().Get("JSON").Call("stringify", result).String()

	var decodeResult DecodeResult
	if err := json.Unmarshal([]byte(resultJSON), &decodeResult); err != nil {
		return nil, fmt.Errorf("failed to parse decode result: %w", err)
	}

	return &decodeResult, nil
}

// DecodeImageFile 从文件路径解码图像
func (r *Runtime) DecodeImageFile(filePath string) (*DecodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// 调用 WASM 函数
	result := r.module.Get("exports").Call("decode_image_file", filePath)

	// 解析结果
	resultJSON := js.Global().Get("JSON").Call("stringify", result).String()

	var decodeResult DecodeResult
	if err := json.Unmarshal([]byte(resultJSON), &decodeResult); err != nil {
		return nil, fmt.Errorf("failed to parse decode result: %w", err)
	}

	return &decodeResult, nil
}

// DecodeMultiple 解码多个条码
func (r *Runtime) DecodeMultiple(imageData []byte, width, height, channels int) ([]*DecodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// 创建 JavaScript 数组
	jsArray := js.Global().Get("Uint8Array").New(len(imageData))
	js.CopyBytesToJS(jsArray, imageData)

	// 调用 WASM 函数
	result := r.module.Get("exports").Call("decode_multiple_barcodes", jsArray, width, height, channels)

	// 解析结果
	resultJSON := js.Global().Get("JSON").Call("stringify", result).String()

	var decodeResults []*DecodeResult
	if err := json.Unmarshal([]byte(resultJSON), &decodeResults); err != nil {
		return nil, fmt.Errorf("failed to parse decode results: %w", err)
	}

	return decodeResults, nil
}

// EncodeText 编码文本为二维码
func (r *Runtime) EncodeText(text string, width, height int) (*EncodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// 调用 WASM 函数
	result := r.module.Get("exports").Call("encode_text_to_qr", text, width, height)

	// 解析结果
	resultJSON := js.Global().Get("JSON").Call("stringify", result).String()

	var encodeResult EncodeResult
	if err := json.Unmarshal([]byte(resultJSON), &encodeResult); err != nil {
		return nil, fmt.Errorf("failed to parse encode result: %w", err)
	}

	return &encodeResult, nil
}

// EncodeTextToBarcode 编码文本为指定格式的条码
func (r *Runtime) EncodeTextToBarcode(text, format string, width, height int) (*EncodeResult, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// 调用 WASM 函数
	result := r.module.Get("exports").Call("encode_text_to_barcode", text, format, width, height)

	// 解析结果
	resultJSON := js.Global().Get("JSON").Call("stringify", result).String()

	var encodeResult EncodeResult
	if err := json.Unmarshal([]byte(resultJSON), &encodeResult); err != nil {
		return nil, fmt.Errorf("failed to parse encode result: %w", err)
	}

	return &encodeResult, nil
}

// GetSupportedFormats 获取支持的条码格式
func (r *Runtime) GetSupportedFormats() ([]string, error) {
	if !r.ready {
		return nil, fmt.Errorf("WASM runtime not initialized")
	}

	// 调用 WASM 函数
	result := r.module.Get("exports").Call("get_supported_formats")

	// 解析结果
	resultJSON := js.Global().Get("JSON").Call("stringify", result).String()

	var formats []string
	if err := json.Unmarshal([]byte(resultJSON), &formats); err != nil {
		return nil, fmt.Errorf("failed to parse formats: %w", err)
	}

	return formats, nil
}

// Close 关闭运行时
func (r *Runtime) Close() error {
	r.ready = false
	r.module = js.Undefined()
	return nil
}
