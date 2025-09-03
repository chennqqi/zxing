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
	fmt.Println("=== ZXing 详细二维码扫描器 ===\n")

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
	fmt.Printf("✅ 图片格式: PNG\n")
	fmt.Println()

	// 使用 CGO 后端进行详细扫描
	fmt.Println("🔍 使用 CGO 后端进行详细扫描...")
	detailedScanWithBackend("cgo", img)

	fmt.Println()

	// 使用 WASM 后端进行详细扫描
	fmt.Println("🔍 使用 WASM 后端进行详细扫描...")
	detailedScanWithBackend("wasm", img)

	fmt.Println("\n=== 详细扫描完成 ===")
}

func detailedScanWithBackend(backend string, img image.Image) {
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

	// 尝试多种解码策略
	decodeStrategies := []struct {
		name string
		opts *zxing.DecodeOptions
	}{
		{
			name: "默认选项",
			opts: &zxing.DecodeOptions{},
		},
		{
			name: "TryHarder 模式",
			opts: &zxing.DecodeOptions{
				TryHarder: true,
			},
		},
		{
			name: "只尝试二维码",
			opts: &zxing.DecodeOptions{
				TryHarder:       true,
				PossibleFormats: []string{"QR_CODE"},
			},
		},
		{
			name: "尝试所有格式",
			opts: &zxing.DecodeOptions{
				TryHarder:       true,
				PossibleFormats: []string{"ALL"},
			},
		},
		{
			name: "指定字符集",
			opts: &zxing.DecodeOptions{
				TryHarder:    true,
				CharacterSet: "UTF-8",
			},
		},
	}

	for i, strategy := range decodeStrategies {
		fmt.Printf("   🔍 策略 %d: %s\n", i+1, strategy.name)

		result, err := zx.DecodeImage(context.Background(), img, strategy.opts)
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

		// 如果成功解码，就不需要尝试其他策略了
		fmt.Printf("      🎯 使用策略 '%s' 成功解码\n", strategy.name)
		break
	}

	// 尝试使用字节数据解码
	fmt.Printf("   🔍 尝试字节数据解码...\n")

	// 将图像转换为字节数据
	imgBytes, width, height, err := imageToBytes(img)
	if err != nil {
		fmt.Printf("      ❌ 图像转字节失败: %v\n", err)
		return
	}

	fmt.Printf("      📊 图像数据: %d bytes, %dx%d\n", len(imgBytes), width, height)

	// 尝试解码字节数据
	result, err := zx.DecodeBytes(context.Background(), imgBytes, width, height, &zxing.DecodeOptions{
		TryHarder:       true,
		PossibleFormats: []string{"QR_CODE"},
	})

	if err != nil {
		fmt.Printf("      ❌ 字节解码失败: %v\n", err)
	} else {
		fmt.Printf("      ✅ 字节解码成功！\n")
		fmt.Printf("         📝 文本内容: %s\n", result.Text)
		fmt.Printf("         🏷️  条码格式: %s\n", result.Format)
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

	return img, nil
}

// imageToBytes 将图像转换为字节数据
func imageToBytes(img image.Image) ([]byte, int, int, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建字节数组
	bytes := make([]byte, width*height*4) // RGBA 格式

	// 填充字节数据
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			idx := (y-bounds.Min.Y)*width*4 + (x-bounds.Min.X)*4
			bytes[idx] = byte(r >> 8)
			bytes[idx+1] = byte(g >> 8)
			bytes[idx+2] = byte(b >> 8)
			bytes[idx+3] = byte(a >> 8)
		}
	}

	return bytes, width, height, nil
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
