package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chennqqi/zxing/pkg/zxing"
	_ "golang.org/x/image/bmp"
)

var (
	version = "1.0.0"
)

func main() {
	var (
		imagePath   = flag.String("i", "", "Image file path to decode")
		imageDir    = flag.String("d", "", "Directory containing images to decode (batch mode)")
		backend     = flag.String("backend", "auto", "Backend to use: auto, cgo, wasm")
		tryHarder   = flag.Bool("try-harder", false, "Try harder to decode")
		formats     = flag.String("formats", "all", "Comma-separated list of formats (QR_CODE, CODE_128, etc.) or 'all'")
		outputJSON  = flag.Bool("json", false, "Output results in JSON format")
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ZXing CLI - Barcode/QR Code Scanner\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -i image.png\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d ./images --try-harder\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -i image.png --backend cgo --formats QR_CODE\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d ./images --json\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("ZXing CLI version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 检查输入
	if *imagePath == "" && *imageDir == "" {
		fmt.Fprintf(os.Stderr, "Error: Either -i (image file) or -d (directory) must be specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// 创建配置
	config := &zxing.Config{
		Backend:  parseBackend(*backend),
		WASMPath: "wasm/zxingwrapper.wasm",
		Debug:    false,
	}

	// 创建ZXing实例
	zx, err := zxing.New(config)
	if err != nil {
		log.Fatalf("Failed to create ZXing instance: %v", err)
	}
	defer zx.Close()

	// 解析格式
	formatList := parseFormats(*formats)

	// 创建解码选项
	decodeOpts := &zxing.DecodeOptions{
		TryHarder:       *tryHarder,
		PossibleFormats: formatList,
	}

	// 处理单个文件或目录
	if *imagePath != "" {
		processImage(zx, *imagePath, decodeOpts, *outputJSON)
	} else if *imageDir != "" {
		processDirectory(zx, *imageDir, decodeOpts, *outputJSON)
	}
}

func parseBackend(backend string) zxing.Backend {
	switch strings.ToLower(backend) {
	case "cgo":
		return zxing.BackendCGO
	case "wasm":
		return zxing.BackendWASM
	case "auto":
		return zxing.BackendAuto
	default:
		log.Printf("Warning: Unknown backend '%s', using 'auto'", backend)
		return zxing.BackendAuto
	}
}

func parseFormats(formats string) []string {
	if formats == "all" || formats == "" {
		return []string{"ALL"}
	}

	parts := strings.Split(formats, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, strings.ToUpper(part))
		}
	}
	return result
}

func processImage(zx zxing.ZXing, imagePath string, opts *zxing.DecodeOptions, jsonOutput bool) {
	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Fatalf("Image file not found: %s", imagePath)
	}

	// 加载图片
	img, err := loadImage(imagePath)
	if err != nil {
		log.Fatalf("Failed to load image: %v", err)
	}

	// 解码
	result, err := zx.DecodeImage(context.Background(), img, opts)
	if err != nil {
		if jsonOutput {
			fmt.Printf(`{"success":false,"error":"%s","file":"%s"}`+"\n", err.Error(), imagePath)
		} else {
			fmt.Printf("❌ Decode failed: %v\n", err)
		}
		os.Exit(1)
	}

	// 输出结果
	if jsonOutput {
		outputJSONResult(imagePath, result)
	} else {
		outputTextResult(imagePath, result)
	}
}

func processDirectory(zx zxing.ZXing, dirPath string, opts *zxing.DecodeOptions, jsonOutput bool) {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Fatalf("Directory not found: %s", dirPath)
	}

	// 读取目录
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	successCount := 0
	failCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 检查是否是图片文件
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !isImageExtension(ext) {
			continue
		}

		imagePath := filepath.Join(dirPath, entry.Name())

		// 加载图片
		img, err := loadImage(imagePath)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s","file":"%s"}`+"\n", err.Error(), imagePath)
			} else {
				fmt.Printf("⚠️  %s: Failed to load image: %v\n", entry.Name(), err)
			}
			failCount++
			continue
		}

		// 解码
		result, err := zx.DecodeImage(context.Background(), img, opts)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s","file":"%s"}`+"\n", err.Error(), imagePath)
			} else {
				fmt.Printf("❌ %s: Decode failed: %v\n", entry.Name(), err)
			}
			failCount++
			continue
		}

		// 输出结果
		if jsonOutput {
			outputJSONResult(imagePath, result)
		} else {
			outputTextResult(imagePath, result)
		}
		successCount++
	}

	if !jsonOutput {
		fmt.Printf("\n=== Summary ===\n")
		fmt.Printf("Success: %d\n", successCount)
		fmt.Printf("Failed: %d\n", failCount)
		fmt.Printf("Total: %d\n", successCount+failCount)
	}
}

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

func isImageExtension(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp":
		return true
	default:
		return false
	}
}

func outputTextResult(imagePath string, result *zxing.Result) {
	fmt.Printf("📷 File: %s\n", imagePath)
	fmt.Printf("✅ Decoded successfully!\n")

	// 显示后端信息
	if backend, ok := result.Metadata["backend"].(string); ok {
		fmt.Printf("   Backend: %s\n", backend)
	}

	fmt.Printf("   Text: %s\n", result.Text)
	fmt.Printf("   Format: %s\n", result.Format)
	if len(result.Points) > 0 {
		fmt.Printf("   Points: %d\n", len(result.Points))
	}
	if len(result.Metadata) > 0 {
		// 已经显示过 backend，跳过
		otherMeta := make(map[string]interface{})
		for k, v := range result.Metadata {
			if k != "backend" {
				otherMeta[k] = v
			}
		}
		if len(otherMeta) > 0 {
			fmt.Printf("   Metadata: %v\n", otherMeta)
		}
	}
	fmt.Println()
}

func outputJSONResult(imagePath string, result *zxing.Result) {
	fmt.Printf(`{"success":true,"file":"%s","text":"%s","format":"%s","points":%d}`+"\n",
		imagePath, result.Text, result.Format, len(result.Points))
}
