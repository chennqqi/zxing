//go:build cgo

package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
	// 创建配置
	config := &zxing.Config{
		Backend: zxing.BackendCGO,
		Debug:   true,
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

	img, err := zx.EncodeText(nil, "Hello, ZXing!", opts)
	if err != nil {
		log.Fatalf("Failed to encode text: %v", err)
	}

	bounds := img.Bounds()
	fmt.Printf("Encoded image: %dx%d\n", bounds.Dx(), bounds.Dy())

	fmt.Println("Test completed successfully!")
}
