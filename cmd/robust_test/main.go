package main

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
	fmt.Println("=== ZXing 健壮的真实数据测试 ===\n")

	// 测试图片路径
	testImagePath := "data/qrcode_www.bing.com.png"

	// 检查测试图片是否存在
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		log.Fatalf("测试图片不存在: %s", testImagePath)
	}

	fmt.Printf("使用测试图片: %s\n", testImagePath)
	fmt.Printf("图片大小: %s\n", formatFileSize(testImagePath))
	fmt.Println()

	// 测试 CGO 后端
	fmt.Println("1. 测试 CGO 后端:")
	testBackend("cgo", testImagePath)

	fmt.Println()

	// 测试 WASM 后端
	fmt.Println("2. 测试 WASM 后端:")
	testBackend("wasm", testImagePath)

	fmt.Println("\n=== 测试完成 ===")
}

func testBackend(backend, imagePath string) {
	fmt.Printf("   使用后端: %s\n", backend)

	// 设置环境变量
	os.Setenv("ZXING_BACKEND", backend)

	// 创建配置
	config := &zxing.Config{
		Backend:  zxing.Backend(backend),
		WASMPath: "wasm/zxing.wasm",
		Debug:    true,
	}

	// 创建 ZXing 实例
	zx, err := zxing.New(config)
	if err != nil {
		fmt.Printf("   ❌ 创建实例失败: %v\n", err)
		return
	}
	defer zx.Close()

	fmt.Printf("   ✅ 实例创建成功，后端: %s\n", zx.GetBackend())

	// 测试编码功能
	fmt.Println("   测试编码功能...")
	encodeOpts := &zxing.EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	img, err := zx.EncodeText(context.Background(), "Hello from "+backend, encodeOpts)
	if err != nil {
		fmt.Printf("   ❌ 编码失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 编码成功，图像尺寸: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())

	// 测试解码刚编码的图片
	fmt.Println("   测试解码刚编码的图片...")
	decodeOpts := &zxing.DecodeOptions{
		TryHarder: true,
	}

	result, err := zx.DecodeImage(context.Background(), img, decodeOpts)
	if err != nil {
		fmt.Printf("   ❌ 解码失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 解码成功，文本: %s，格式: %s\n", result.Text, result.Format)

	// 尝试解码真实图片
	fmt.Println("   尝试解码真实图片...")
	realImg, err := loadImageRobust(imagePath)
	if err != nil {
		fmt.Printf("   ❌ 加载真实图片失败: %v\n", err)
		return
	}

	fmt.Printf("   ✅ 图片加载成功，尺寸: %dx%d\n", realImg.Bounds().Dx(), realImg.Bounds().Dy())

	realResult, err := zx.DecodeImage(context.Background(), realImg, decodeOpts)
	if err != nil {
		fmt.Printf("   ❌ 解码真实图片失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 解码真实图片成功！\n")
	fmt.Printf("      文本内容: %s\n", realResult.Text)
	fmt.Printf("      条码格式: %s\n", realResult.Format)
	fmt.Printf("      位置点数量: %d\n", len(realResult.Points))

	if len(realResult.Metadata) > 0 {
		fmt.Printf("      元数据: %v\n", realResult.Metadata)
	}
}

// loadImageRobust 健壮的图片加载函数，支持多种格式
func loadImageRobust(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 尝试解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("解码图片失败 (格式: %s): %v", format, err)
	}

	fmt.Printf("      检测到图片格式: %s\n", format)
	return img, nil
}

// formatFileSize 格式化文件大小
func formatFileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "未知"
	}

	size := info.Size()
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
}
