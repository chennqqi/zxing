package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/zxing"
)

func main() {
	// 创建默认选项
	options := zxing.NewDefaultOptions()
	if options == nil {
		log.Fatal("Failed to create default options")
	}

	// 设置只识别二维码
	options.Formats = zxing.FormatQRCode

	// 解码单个二维码
	result, err := zxing.Decode("test.png", options)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	fmt.Printf("Decoded text: %s\n", result.Text)
	fmt.Printf("Format: %v\n", result.Format)
	fmt.Printf("Confidence: %.2f\n", result.Confidence)

	// 解码多个条码
	options.Formats = zxing.FormatAll // 支持所有格式
	results, err := zxing.DecodeMulti("test.png", options)
	if err != nil {
		log.Fatalf("Failed to decode multiple: %v", err)
	}

	fmt.Printf("\nFound %d barcodes:\n", len(results))
	for i, result := range results {
		fmt.Printf("\nBarcode %d:\n", i+1)
		fmt.Printf("  Text: %s\n", result.Text)
		fmt.Printf("  Format: %v\n", result.Format)
		fmt.Printf("  Confidence: %.2f\n", result.Confidence)
	}
}
