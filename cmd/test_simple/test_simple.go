package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/zxing"
)

func main() {
	fmt.Println("=== ZXing Go Wrapper 测试 ===")
	
	// 创建默认解码选项
	options := zxing.NewDefaultOptions()
	if options == nil {
		log.Fatal("创建默认选项失败")
	}
	
	fmt.Printf("默认选项: Formats=%v, TryHarder=%v, TryRotate=%v\n", 
		options.Formats, options.TryHarder, options.TryRotate)
	
	// 测试第一个图片
	fmt.Println("\n--- 测试图片1: 0ee172907feade36d28b35175ab59fb881de151df42f5801bd2412f2cc425e67.jpg ---")
	result1, err := zxing.Decode("tests/0ee172907feade36d28b35175ab59fb881de151df42f5801bd2412f2cc425e67.jpg", options)
	if err != nil {
		fmt.Printf("解码失败: %v\n", err)
	} else {
		fmt.Printf("解码成功: 文本=%s, 格式=%v, 置信度=%.2f\n", 
			result1.Text, result1.Format, result1.Confidence)
	}
	
	// 测试第二个图片
	fmt.Println("\n--- 测试图片2: 工资补贴领取.jpg ---")
	result2, err := zxing.Decode("tests/工资补贴领取.jpg", options)
	if err != nil {
		fmt.Printf("解码失败: %v\n", err)
	} else {
		fmt.Printf("解码成功: 文本=%s, 格式=%v, 置信度=%.2f\n", 
			result2.Text, result2.Format, result2.Confidence)
	}
	
	// 测试多码解码
	fmt.Println("\n--- 测试多码解码 ---")
	results, err := zxing.DecodeMulti("tests/0ee172907feade36d28b35175ab59fb881de151df42f5801bd2412f2cc425e67.jpg", options)
	if err != nil {
		fmt.Printf("多码解码失败: %v\n", err)
	} else {
		fmt.Printf("找到 %d 个条码:\n", len(results))
		for i, r := range results {
			fmt.Printf("  %d. 文本=%s, 格式=%v, 置信度=%.2f\n", 
				i+1, r.Text, r.Format, r.Confidence)
		}
	}
	
	fmt.Println("\n=== 测试完成 ===")
} 