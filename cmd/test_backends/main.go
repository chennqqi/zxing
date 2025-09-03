package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
	fmt.Println("=== ZXing 后端切换测试 ===\n")

	// 测试 CGO 后端
	fmt.Println("1. 测试 CGO 后端:")
	if err := testBackend("cgo"); err != nil {
		log.Printf("CGO 后端测试失败: %v", err)
	} else {
		fmt.Println("✅ CGO 后端测试成功")
	}

	fmt.Println()

	// 测试 WASM 后端
	fmt.Println("2. 测试 WASM 后端:")
	if err := testBackend("wasm"); err != nil {
		log.Printf("WASM 后端测试失败: %v", err)
	} else {
		fmt.Println("✅ WASM 后端测试成功")
	}

	fmt.Println("\n=== 测试完成 ===")
}

func testBackend(backend string) error {
	// 设置环境变量
	os.Setenv("ZXING_BACKEND", backend)
	
	// 创建配置
	config := &zxing.Config{
		Backend:  zxing.Backend(backend),
		WASMPath: "../../wasm/zxing.wasm",
		Debug:    true,
	}

	// 创建 ZXing 实例
	zx, err := zxing.New(config)
	if err != nil {
		return fmt.Errorf("创建实例失败: %v", err)
	}
	defer zx.Close()

	fmt.Printf("   使用后端: %s\n", zx.GetBackend())

	// 测试编码功能
	fmt.Println("   测试编码功能...")
	encodeOpts := &zxing.EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	img, err := zx.EncodeText(nil, "Hello, ZXing "+backend+"!", encodeOpts)
	if err != nil {
		return fmt.Errorf("编码失败: %v", err)
	}
	fmt.Printf("   编码成功，图像尺寸: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())

	// 测试解码功能
	fmt.Println("   测试解码功能...")
	decodeOpts := &zxing.DecodeOptions{
		TryHarder: true,
	}

	result, err := zx.DecodeImage(nil, img, decodeOpts)
	if err != nil {
		return fmt.Errorf("解码失败: %v", err)
	}
	fmt.Printf("   解码成功，文本: %s，格式: %s\n", result.Text, result.Format)

	return nil
}
