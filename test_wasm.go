//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"syscall/js"

	"github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
	// 创建配置
	config := &zxing.Config{
		Backend:  zxing.BackendWASM,
		WASMPath: "./wasm/zxingwrapper.wasm",
		Debug:    true,
	}

	// 创建ZXing实例
	zx, err := zxing.New(config)
	if err != nil {
		log.Fatalf("Failed to create ZXing instance: %v", err)
	}
	defer zx.Close()

	fmt.Printf("ZXing backend: %s\n", zx.GetBackend())

	// 测试编码
	opts := &zxing.EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	img, err := zx.EncodeText(nil, "Hello, ZXing WASM!", opts)
	if err != nil {
		log.Fatalf("Failed to encode text: %v", err)
	}

	bounds := img.Bounds()
	fmt.Printf("Encoded image: %dx%d\n", bounds.Dx(), bounds.Dy())

	// 测试解码（模拟图像数据）
	testImageData := make([]byte, 256*256*4) // RGBA格式
	for i := range testImageData {
		testImageData[i] = byte(i % 256)
	}

	result, err := zx.DecodeBytes(nil, testImageData, 256, 256, &zxing.DecodeOptions{
		TryHarder: true,
	})
	if err != nil {
		fmt.Printf("Decode test (expected to fail): %v\n", err)
	} else {
		fmt.Printf("Decoded text: %s\n", result.Text)
	}

	fmt.Println("WASM test completed successfully!")

	// 保持程序运行（在WASM环境中需要）
	select {}
}
