package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/zxing"
)

func main() {
	// 创建默认解码选项
	opts := zxing.NewDefaultOptions()
	if opts == nil {
		log.Fatal("Failed to create default options")
	}

	// 解码单个二维码
	result, err := zxing.Decode("testdata/qrcode.png", opts)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	fmt.Printf("Decoded text: %s\n", result.Text)
	fmt.Printf("Format: %s\n", result.Format)
	fmt.Printf("Confidence: %.2f\n", result.Confidence)

	// 解码多个条码
	results, err := zxing.DecodeMulti("testdata/multi.png", opts)
	if err != nil {
		log.Fatalf("Failed to decode multiple: %v", err)
	}

	fmt.Printf("\nFound %d barcodes:\n", len(results))
	for i, r := range results {
		fmt.Printf("%d. Text: %s, Format: %s, Confidence: %.2f\n",
			i+1, r.Text, r.Format, r.Confidence)
	}
}
