package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/threatbook/zxing"
)

func main() {
	// 创建自定义解码选项
	opts := zxing.NewDefaultOptions()
	if opts == nil {
		log.Fatal("Failed to create default options")
	}

	// 设置只解码QR码
	opts.Formats = zxing.FormatQRCode

	// 设置尝试次数
	opts.TryHarder = true
	opts.TryRotate = true

	// 设置最小置信度
	opts.MinConfidence = 0.8

	// 解码目录中的所有图片
	dir := "testdata"
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			path := filepath.Join(dir, file.Name())
			fmt.Printf("\nProcessing %s:\n", path)

			// 尝试解码单个条码
			result, err := zxing.Decode(path, opts)
			if err != nil {
				fmt.Printf("Single decode failed: %v\n", err)
			} else {
				fmt.Printf("Single decode result:\n")
				fmt.Printf("  Text: %s\n", result.Text)
				fmt.Printf("  Format: %s\n", result.Format)
				fmt.Printf("  Confidence: %.2f\n", result.Confidence)
			}

			// 尝试解码多个条码
			results, err := zxing.DecodeMulti(path, opts)
			if err != nil {
				fmt.Printf("Multi decode failed: %v\n", err)
			} else {
				fmt.Printf("Multi decode results (%d found):\n", len(results))
				for i, r := range results {
					fmt.Printf("  %d. Text: %s, Format: %s, Confidence: %.2f\n",
						i+1, r.Text, r.Format, r.Confidence)
				}
			}
		}
	}
}

func isImageFile(name string) bool {
	ext := filepath.Ext(name)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp":
		return true
	default:
		return false
	}
}
