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
	fmt.Println("=== ZXing 二维码扫描器 ===\n")

	// 测试图片路径
	testImagePath := "data/qrcode_www.bing.com.png"
	
	// 检查测试图片是否存在
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		log.Fatalf("测试图片不存在: %s", testImagePath)
	}

	fmt.Printf("📷 扫描图片: %s\n", testImagePath)
	fmt.Printf("📏 图片大小: %s\n", formatFileSize(testImagePath))
	fmt.Println()

	// 加载图片
	fmt.Println("🔄 加载图片...")
	img, err := loadImage(testImagePath)
	if err != nil {
		log.Fatalf("加载图片失败: %v", err)
	}
	fmt.Printf("✅ 图片加载成功，尺寸: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Println()

	// 使用 CGO 后端扫描
	fmt.Println("🔍 使用 CGO 后端扫描...")
	scanWithBackend("cgo", img)

	fmt.Println()

	// 使用 WASM 后端扫描
	fmt.Println("🔍 使用 WASM 后端扫描...")
	scanWithBackend("wasm", img)

	fmt.Println("\n=== 扫描完成 ===")
}

func scanWithBackend(backend string, img image.Image) {
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
		fmt.Printf("❌ 创建实例失败: %v\n", err)
		return
	}
	defer zx.Close()

	fmt.Printf("   🚀 后端: %s\n", zx.GetBackend())

	// 尝试多种解码选项
	decodeOptions := []*zxing.DecodeOptions{
		{TryHarder: false},                    // 默认选项
		{TryHarder: true},                     // 更努力地解码
		{TryHarder: true, PossibleFormats: []string{"QR_CODE"}}, // 只尝试二维码
		{TryHarder: true, PossibleFormats: []string{"ALL"}},    // 尝试所有格式
	}

	for i, opts := range decodeOptions {
		fmt.Printf("   🔍 尝试解码选项 %d...\n", i+1)
		
		result, err := zx.DecodeImage(context.Background(), img, opts)
		if err != nil {
			fmt.Printf("      ❌ 解码失败: %v\n", err)
			continue
		}

		fmt.Printf("      ✅ 解码成功！\n")
		fmt.Printf("         📝 文本内容: %s\n", result.Text)
		fmt.Printf("         🏷️  条码格式: %s\n", result.Format)
		fmt.Printf("         📍 位置点数量: %d\n", len(result.Points))
		
		if len(result.Points) > 0 {
			fmt.Printf("         📍 位置点: %v\n", result.Points)
		}
		
		if len(result.Metadata) > 0 {
			fmt.Printf("         📊 元数据: %v\n", result.Metadata)
		}
		
		// 如果成功解码，就不需要尝试其他选项了
		break
	}
}

// loadImage 加载图片文件
func loadImage(path string) (image.Image, error) {
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

	fmt.Printf("✅ 检测到图片格式: %s\n", format)
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
