// WASM 集成示例程序
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chennqqi/zxing/pkg/zxing"
)

func main() {
	// 创建配置
	config := zxing.LoadConfigFromEnv()
	config.Debug = true

	// 打印配置信息
	fmt.Printf("使用配置: Backend=%s, WASMPath=%s, Debug=%t\n",
		config.Backend, config.WASMPath, config.Debug)

	// 创建 ZXing 实例
	zx, err := zxing.New(config)
	if err != nil {
		log.Fatalf("创建 ZXing 实例失败: %v", err)
	}
	defer zx.Close()

	fmt.Printf("使用后端: %s\n", zx.GetBackend())

	// 测试编码功能
	fmt.Println("\n=== 测试编码功能 ===")
	testEncode(zx)

	// 测试解码功能
	fmt.Println("\n=== 测试解码功能 ===")
	testDecode(zx)
}

// testEncode 测试编码功能
func testEncode(zx zxing.ZXing) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	text := "Hello, ZXing WASM!"
	opts := &zxing.EncodeOptions{
		Width:  256,
		Height: 256,
		Format: "QR_CODE",
	}

	fmt.Printf("编码文本: %s\n", text)

	// 编码为图像
	img, err := zx.EncodeText(ctx, text, opts)
	if err != nil {
		fmt.Printf("编码失败: %v\n", err)
		return
	}

	bounds := img.Bounds()
	fmt.Printf("生成图像尺寸: %dx%d\n", bounds.Dx(), bounds.Dy())

	// 编码为字节数据
	data, width, height, err := zx.EncodeToBytes(ctx, text, opts)
	if err != nil {
		fmt.Printf("编码为字节失败: %v\n", err)
		return
	}

	fmt.Printf("生成字节数据: %d bytes, 尺寸: %dx%d\n", len(data), width, height)
}

// testDecode 测试解码功能
func testDecode(zx zxing.ZXing) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建测试数据（简单的黑白图像）
	width, height := 256, 256
	data := make([]byte, width*height*4) // RGBA

	// 填充测试图案
	for i := 0; i < len(data); i += 4 {
		// 简单的棋盘图案
		x := (i / 4) % width
		y := (i / 4) / width

		if (x/32+y/32)%2 == 0 {
			data[i] = 0     // R
			data[i+1] = 0   // G
			data[i+2] = 0   // B
			data[i+3] = 255 // A
		} else {
			data[i] = 255   // R
			data[i+1] = 255 // G
			data[i+2] = 255 // B
			data[i+3] = 255 // A
		}
	}

	opts := &zxing.DecodeOptions{
		TryHarder: true,
	}

	fmt.Printf("解码图像数据: %d bytes, 尺寸: %dx%d\n", len(data), width, height)

	result, err := zx.DecodeBytes(ctx, data, width, height, opts)
	if err != nil {
		fmt.Printf("解码失败: %v\n", err)
		return
	}

	fmt.Printf("解码结果: %s (格式: %s)\n", result.Text, result.Format)
	fmt.Printf("位置点数量: %d\n", len(result.Points))
}

// init 初始化函数
func init() {
	// 设置默认环境变量（如果未设置）
	if os.Getenv("ZXING_BACKEND") == "" {
		os.Setenv("ZXING_BACKEND", "wasm")
	}

	if os.Getenv("ZXING_WASM_PATH") == "" {
		os.Setenv("ZXING_WASM_PATH", "../../wasm/zxingwrapper.wasm")
	}
}
