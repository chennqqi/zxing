package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
	fmt.Println("=== ZXing 真实数据测试 ===\n")

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
	if err := testBackendWithRealData("cgo", testImagePath); err != nil {
		log.Printf("CGO 后端测试失败: %v", err)
	} else {
		fmt.Println("✅ CGO 后端测试成功")
	}

	fmt.Println()

	// 测试 WASM 后端
	fmt.Println("2. 测试 WASM 后端:")
	if err := testBackendWithRealData("wasm", testImagePath); err != nil {
		log.Printf("WASM 后端测试失败: %v", err)
	} else {
		fmt.Println("✅ WASM 后端测试成功")
	}

	fmt.Println("\n=== 真实数据测试完成 ===")
}

func testBackendWithRealData(backend, imagePath string) error {
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

	// 测试解码真实图片
	fmt.Println("   测试解码真实图片...")
	
	// 读取图片文件
	img, err := loadImage(imagePath)
	if err != nil {
		return fmt.Errorf("加载图片失败: %v", err)
	}

	// 解码图片
	decodeOpts := &zxing.DecodeOptions{
		TryHarder: true,
	}

	result, err := zx.DecodeImage(context.Background(), img, decodeOpts)
	if err != nil {
		return fmt.Errorf("解码失败: %v", err)
	}

	fmt.Printf("   解码成功！\n")
	fmt.Printf("   文本内容: %s\n", result.Text)
	fmt.Printf("   条码格式: %s\n", result.Format)
	fmt.Printf("   位置点数量: %d\n", len(result.Points))
	
	if len(result.Metadata) > 0 {
		fmt.Printf("   元数据: %v\n", result.Metadata)
	}

	// 测试编码功能
	fmt.Println("   测试编码功能...")
	encodeOpts := &zxing.EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	encodedImg, err := zx.EncodeText(context.Background(), "Test from "+backend+" backend", encodeOpts)
	if err != nil {
		return fmt.Errorf("编码失败: %v", err)
	}
	fmt.Printf("   编码成功，图像尺寸: %dx%d\n", encodedImg.Bounds().Dx(), encodedImg.Bounds().Dy())

	// 测试解码刚编码的图片
	fmt.Println("   测试解码刚编码的图片...")
	decodedResult, err := zx.DecodeImage(context.Background(), encodedImg, decodeOpts)
	if err != nil {
		return fmt.Errorf("解码编码后的图片失败: %v", err)
	}
	fmt.Printf("   解码编码后的图片成功，文本: %s\n", decodedResult.Text)

	return nil
}

// loadImage 加载图片文件
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

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
