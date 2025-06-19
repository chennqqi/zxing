package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	_ "image/png"
	_ "image/gif"
	_ "image/color"
	_ "image/draw"
	_ "image/jpeg"

	"github.com/chennqqi/zxing"
	gozxing "github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// 图片格式的魔数
var imageMagicNumbers = map[string][]byte{
	"jpeg": {0xFF, 0xD8, 0xFF},
	"png":  {0x89, 0x50, 0x4E, 0x47},
	"gif":  {0x47, 0x49, 0x46},
	"bmp":  {0x42, 0x4D},
}

// detectImageFormat 通过文件头检测图片格式
func detectImageFormat(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取文件头
	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		return "", err
	}

	// 检查各种格式的魔数
	for format, magic := range imageMagicNumbers {
		if bytes.HasPrefix(header, magic) {
			return format, nil
		}
	}

	return "", fmt.Errorf("未知的图片格式")
}

// isImageFile 检查文件是否为图片
func isImageFile(filePath string) bool {
	format, err := detectImageFormat(filePath)
	if err != nil {
		return false
	}
	return format != ""
}

// CompareResult 表示对比测试结果
type CompareResult struct {
	ImagePath     string
	ImageFormat   string
	OurResult     *zxing.DecodeResult
	OurError      string
	OurTime       time.Duration
	GozxingResult *gozxing.Result
	GozxingError  string
	GozxingTime   time.Duration
	Success       bool
}

// CompareReport 表示对比测试报告
type CompareReport struct {
	Timestamp       time.Time
	TotalImages     int
	OurSuccessCount int
	GozxingSuccessCount int
	BothSuccessCount    int
	OurOnlySuccessCount int
	GozxingOnlySuccessCount int
	BothFailedCount     int
	Results             []CompareResult
}

func main() {
	// 设置动态库路径
	os.Setenv("LD_LIBRARY_PATH", "./lib")

	// 测试图片目录
	testDir := "images"
	
	// 创建对比报告
	report := &CompareReport{
		Timestamp: time.Now(),
		Results:   []CompareResult{},
	}

	fmt.Println("开始ZXing库对比测试...")
	fmt.Printf("测试目录: %s\n", testDir)
	fmt.Println(strings.Repeat("=", 60))

	// 获取所有文件
	files, err := os.ReadDir(testDir)
	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	// 过滤出图片文件
	var imageFiles []string
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(testDir, file.Name())
			if isImageFile(filePath) {
				imageFiles = append(imageFiles, filePath)
			}
		}
	}

	report.TotalImages = len(imageFiles)
	fmt.Printf("找到 %d 个图片文件\n", report.TotalImages)

	if report.TotalImages == 0 {
		fmt.Println("未找到有效的图片文件")
		return
	}

	// 创建默认选项
	options := zxing.NewDefaultOptions()
	if options == nil {
		log.Fatal("创建默认选项失败")
	}

	// 测试每个图片
	for i, imagePath := range imageFiles {
		format, _ := detectImageFormat(imagePath)
		fmt.Printf("\n测试图片 %d/%d: %s (格式: %s)\n", i+1, report.TotalImages, filepath.Base(imagePath), format)
		
		result := CompareResult{
			ImagePath:   imagePath,
			ImageFormat: format,
		}

		// 测试我们的ZXing wrapper
		fmt.Println("  测试我们的ZXing wrapper...")
		startTime := time.Now()
		ourResult, ourErr := zxing.Decode(imagePath, options)
		ourTime := time.Since(startTime)
		
		result.OurTime = ourTime
		if ourErr != nil {
			result.OurError = ourErr.Error()
			fmt.Printf("    失败: %v\n", ourErr)
		} else {
			result.OurResult = ourResult
			fmt.Printf("    成功: %s (格式: %s, 耗时: %v)\n", 
				ourResult.Text, ourResult.Format.String(), ourTime)
		}

		// 测试gozxing
		fmt.Println("  测试gozxing...")
		startTime = time.Now()
		gozxingResult, gozxingErr := testGozxing(imagePath)
		gozxingTime := time.Since(startTime)
		
		result.GozxingTime = gozxingTime
		if gozxingErr != nil {
			result.GozxingError = gozxingErr.Error()
			fmt.Printf("    失败: %v\n", gozxingErr)
		} else {
			result.GozxingResult = gozxingResult
			fmt.Printf("    成功: %s (格式: %s, 耗时: %v)\n", 
				gozxingResult.GetText(), gozxingResult.GetBarcodeFormat().String(), gozxingTime)
		}

		// 统计结果
		ourSuccess := result.OurResult != nil
		gozxingSuccess := result.GozxingResult != nil
		
		if ourSuccess && gozxingSuccess {
			report.BothSuccessCount++
			result.Success = true
		} else if ourSuccess {
			report.OurOnlySuccessCount++
			result.Success = true
		} else if gozxingSuccess {
			report.GozxingOnlySuccessCount++
			result.Success = true
		} else {
			report.BothFailedCount++
		}
		
		if ourSuccess {
			report.OurSuccessCount++
		}
		if gozxingSuccess {
			report.GozxingSuccessCount++
		}

		report.Results = append(report.Results, result)
	}

	// 生成对比报告
	generateCompareReport(report)
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("对比测试完成！详细报告已保存到 compare_report.txt")
}

func testGozxing(imagePath string) (*gozxing.Result, error) {
	// 打开图片文件
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("打开图片失败: %v", err)
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %v", err)
	}

	// 创建二维码读取器
	reader := qrcode.NewQRCodeReader()
	
	// 创建二进制位图
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("创建二进制位图失败: %v", err)
	}

	// 读取二维码
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return nil, fmt.Errorf("解码失败: %v", err)
	}

	return result, nil
}

func generateCompareReport(report *CompareReport) {
	file, err := os.Create("compare_report.txt")
	if err != nil {
		log.Fatalf("创建报告文件失败: %v", err)
	}
	defer file.Close()

	// 写入报告头部
	fmt.Fprintf(file, "ZXing库对比测试报告\n")
	fmt.Fprintf(file, "生成时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 60))

	// 写入统计信息
	fmt.Fprintf(file, "测试统计:\n")
	fmt.Fprintf(file, "  总图片数: %d\n", report.TotalImages)
	fmt.Fprintf(file, "  我们的ZXing成功数: %d (%.2f%%)\n", 
		report.OurSuccessCount, float64(report.OurSuccessCount)/float64(report.TotalImages)*100)
	fmt.Fprintf(file, "  Gozxing成功数: %d (%.2f%%)\n", 
		report.GozxingSuccessCount, float64(report.GozxingSuccessCount)/float64(report.TotalImages)*100)
	fmt.Fprintf(file, "  两者都成功: %d\n", report.BothSuccessCount)
	fmt.Fprintf(file, "  仅我们的成功: %d\n", report.OurOnlySuccessCount)
	fmt.Fprintf(file, "  仅Gozxing成功: %d\n", report.GozxingOnlySuccessCount)
	fmt.Fprintf(file, "  两者都失败: %d\n", report.BothFailedCount)
	fmt.Fprintf(file, "\n")

	// 写入详细结果
	fmt.Fprintf(file, "详细结果:\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("=", 60))
	
	for i, result := range report.Results {
		fmt.Fprintf(file, "\n%d. 图片: %s (格式: %s)\n", i+1, filepath.Base(result.ImagePath), result.ImageFormat)
		fmt.Fprintf(file, "   路径: %s\n", result.ImagePath)
		
		// 我们的结果
		fmt.Fprintf(file, "   我们的ZXing:\n")
		if result.OurResult != nil {
			fmt.Fprintf(file, "     成功: %s\n", result.OurResult.Text)
			fmt.Fprintf(file, "     格式: %s\n", result.OurResult.Format.String())
			fmt.Fprintf(file, "     置信度: %.2f\n", result.OurResult.Confidence)
			fmt.Fprintf(file, "     耗时: %v\n", result.OurTime)
		} else {
			fmt.Fprintf(file, "     失败: %s\n", result.OurError)
			fmt.Fprintf(file, "     耗时: %v\n", result.OurTime)
		}
		
		// Gozxing结果
		fmt.Fprintf(file, "   Gozxing:\n")
		if result.GozxingResult != nil {
			fmt.Fprintf(file, "     成功: %s\n", result.GozxingResult.GetText())
			fmt.Fprintf(file, "     格式: %s\n", result.GozxingResult.GetBarcodeFormat().String())
			fmt.Fprintf(file, "     耗时: %v\n", result.GozxingTime)
		} else {
			fmt.Fprintf(file, "     失败: %s\n", result.GozxingError)
			fmt.Fprintf(file, "     耗时: %v\n", result.GozxingTime)
		}
		
		// 对比分析
		if result.OurResult != nil && result.GozxingResult != nil {
			if result.OurResult.Text == result.GozxingResult.GetText() {
				fmt.Fprintf(file, "   结果对比: 一致 ✓\n")
			} else {
				fmt.Fprintf(file, "   结果对比: 不一致 ✗\n")
				fmt.Fprintf(file, "     我们的: %s\n", result.OurResult.Text)
				fmt.Fprintf(file, "     Gozxing: %s\n", result.GozxingResult.GetText())
			}
			
			if result.OurTime < result.GozxingTime {
				fmt.Fprintf(file, "   性能对比: 我们的更快 (%.2fx)\n", float64(result.GozxingTime)/float64(result.OurTime))
			} else {
				fmt.Fprintf(file, "   性能对比: Gozxing更快 (%.2fx)\n", float64(result.OurTime)/float64(result.GozxingTime))
			}
		}
	}

	// 写入性能分析
	fmt.Fprintf(file, "\n性能分析:\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("=", 60))
	
	var ourTotalTime, gozxingTotalTime time.Duration
	var ourSuccessCount, gozxingSuccessCount int
	
	for _, result := range report.Results {
		if result.OurResult != nil {
			ourTotalTime += result.OurTime
			ourSuccessCount++
		}
		if result.GozxingResult != nil {
			gozxingTotalTime += result.GozxingTime
			gozxingSuccessCount++
		}
	}
	
	if ourSuccessCount > 0 {
		fmt.Fprintf(file, "我们的ZXing平均解码时间: %v\n", ourTotalTime/time.Duration(ourSuccessCount))
	}
	if gozxingSuccessCount > 0 {
		fmt.Fprintf(file, "Gozxing平均解码时间: %v\n", gozxingTotalTime/time.Duration(gozxingSuccessCount))
	}
	
	// 写入总结
	fmt.Fprintf(file, "\n总结:\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("=", 60))
	
	if report.OurSuccessCount > report.GozxingSuccessCount {
		fmt.Fprintf(file, "我们的ZXing wrapper在识别成功率上更优\n")
	} else if report.GozxingSuccessCount > report.OurSuccessCount {
		fmt.Fprintf(file, "Gozxing在识别成功率上更优\n")
	} else {
		fmt.Fprintf(file, "两个库在识别成功率上相当\n")
	}
	
	if ourTotalTime < gozxingTotalTime {
		fmt.Fprintf(file, "我们的ZXing wrapper在性能上更优\n")
	} else if gozxingTotalTime < ourTotalTime {
		fmt.Fprintf(file, "Gozxing在性能上更优\n")
	} else {
		fmt.Fprintf(file, "两个库在性能上相当\n")
	}
} 