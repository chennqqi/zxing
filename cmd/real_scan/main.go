package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chennqqi/zxing"
)

func main() {
	fmt.Println("=== ZXing 真实 CGO 接口扫描器 ===\n")

	// 测试图片路径
	testImagePath := "data/qrcode_www.bing.com.png"
	
	// 检查测试图片是否存在
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		log.Fatalf("测试图片不存在: %s", testImagePath)
	}

	fmt.Printf("📷 扫描图片: %s\n", testImagePath)
	fmt.Printf("📏 图片大小: %s\n", formatFileSize(testImagePath))
	fmt.Println()

	// 创建默认解码选项
	fmt.Println("🔧 创建解码选项...")
	options := zxing.NewDefaultOptions()
	if options == nil {
		log.Fatal("Failed to create default options")
	}

	// 设置只识别二维码
	options.Formats = zxing.FormatQRCode
	options.TryHarder = true
	options.TryRotate = true

	fmt.Printf("✅ 解码选项创建成功\n")
	fmt.Printf("   📋 格式: %s\n", options.Formats.String())
	fmt.Printf("   🔍 TryHarder: %t\n", options.TryHarder)
	fmt.Printf("   🔄 TryRotate: %t\n", options.TryRotate)
	fmt.Printf("   🔃 TryInvert: %t\n", options.TryInvert)
	fmt.Printf("   📉 TryDownscale: %t\n", options.TryDownscale)
	fmt.Println()

	// 尝试解码
	fmt.Println("🔍 开始解码二维码...")
	result, err := zxing.Decode(testImagePath, options)
	if err != nil {
		log.Fatalf("解码失败: %v", err)
	}

	fmt.Println("✅ 解码成功！")
	fmt.Printf("📝 文本内容: %s\n", result.Text)
	fmt.Printf("🏷️  条码格式: %s\n", result.Format.String())
	fmt.Printf("📊 置信度: %.2f\n", result.Confidence)

	// 尝试解码多个条码
	fmt.Println("\n🔍 尝试解码多个条码...")
	results, err := zxing.DecodeMulti(testImagePath, options)
	if err != nil {
		fmt.Printf("⚠️  多条码解码失败: %v\n", err)
	} else {
		fmt.Printf("✅ 发现 %d 个条码:\n", len(results))
		for i, res := range results {
			fmt.Printf("   %d. 文本: %s, 格式: %s, 置信度: %.2f\n", 
				i+1, res.Text, res.Format.String(), res.Confidence)
		}
	}

	fmt.Println("\n=== 真实扫描完成 ===")
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
