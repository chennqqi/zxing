package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chennqqi/zxing"
)

func main() {
	dir := "../../tests"
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		os.Exit(1)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !(filepath.Ext(name) == ".jpg" || filepath.Ext(name) == ".png" || filepath.Ext(name) == ".jpeg" || filepath.Ext(name) == ".bmp") {
			continue
		}
		imgPath := filepath.Join(dir, name)
		fmt.Printf("\n==== 测试文件: %s ===="+"\n", imgPath)
		options := zxing.NewDefaultOptions()
		if options == nil {
			fmt.Println("创建解码选项失败")
			continue
		}
		result, err := zxing.Decode(imgPath, options)
		if err != nil {
			fmt.Printf("解码失败: %v\n", err)
			continue
		}
		fmt.Printf("解码内容: %s\n格式: %v\n置信度: %f\n", result.Text, result.Format, result.Confidence)
	}
} 