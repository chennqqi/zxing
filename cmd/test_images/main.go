package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chennqqi/zxing"
)

// TestResult 表示单个测试结果
type TestResult struct {
	ImagePath    string
	Success      bool
	Error        string
	DecodeTime   time.Duration
	Result       *zxing.DecodeResult
	MultiResults []*zxing.DecodeResult
}

// TestReport 表示测试报告
type TestReport struct {
	Timestamp    time.Time
	TotalImages  int
	SuccessCount int
	FailedCount  int
	TotalTime    time.Duration
	Results      []TestResult
}

func main() {
	// 设置动态库路径
	os.Setenv("LD_LIBRARY_PATH", "./lib")

	// 测试图片目录
	testDir := "tests"
	
	// 创建测试报告
	report := &TestReport{
		Timestamp: time.Now(),
		Results:   []TestResult{},
	}

	fmt.Println("开始ZXing Go Wrapper测试...")
	fmt.Printf("测试目录: %s\n", testDir)
	fmt.Println(strings.Repeat("=", 50))

	// 获取所有图片文件
	imageFiles, err := filepath.Glob(filepath.Join(testDir, "*.jpg"))
	if err != nil {
		log.Fatalf("获取图片文件失败: %v", err)
	}

	pngFiles, _ := filepath.Glob(filepath.Join(testDir, "*.png"))
	jpegFiles, _ := filepath.Glob(filepath.Join(testDir, "*.jpeg"))
	
	imageFiles = append(imageFiles, pngFiles...)
	imageFiles = append(imageFiles, jpegFiles...)

	report.TotalImages = len(imageFiles)
	fmt.Printf("找到 %d 个图片文件\n", report.TotalImages)

	// 创建默认选项
	options := zxing.NewDefaultOptions()
	if options == nil {
		log.Fatal("创建默认选项失败")
	}

	// 测试每个图片
	for _, imagePath := range imageFiles {
		fmt.Printf("\n测试图片: %s\n", filepath.Base(imagePath))
		
		result := TestResult{
			ImagePath: imagePath,
		}

		// 测试单个条码解码
		startTime := time.Now()
		decodeResult, err := zxing.Decode(imagePath, options)
		decodeTime := time.Since(startTime)
		
		result.DecodeTime = decodeTime
		
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			fmt.Printf("  单条码解码失败: %v\n", err)
		} else {
			result.Success = true
			result.Result = decodeResult
			fmt.Printf("  单条码解码成功: %s (格式: %s, 置信度: %.2f)\n", 
				decodeResult.Text, decodeResult.Format.String(), decodeResult.Confidence)
		}

		// 测试多个条码解码
		startTime = time.Now()
		multiResults, err := zxing.DecodeMulti(imagePath, options)
		multiDecodeTime := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("  多条码解码失败: %v\n", err)
		} else {
			result.MultiResults = multiResults
			fmt.Printf("  多条码解码成功: 找到 %d 个条码\n", len(multiResults))
			for i, res := range multiResults {
				fmt.Printf("    条码 %d: %s (格式: %s, 置信度: %.2f)\n", 
					i+1, res.Text, res.Format.String(), res.Confidence)
			}
		}

		// 统计结果
		if result.Success {
			report.SuccessCount++
		} else {
			report.FailedCount++
		}
		report.TotalTime += decodeTime + multiDecodeTime

		report.Results = append(report.Results, result)
	}

	// 生成测试报告
	generateReport(report)
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("测试完成！详细报告已保存到 test_report.txt")
}

func generateReport(report *TestReport) {
	file, err := os.Create("test_report.txt")
	if err != nil {
		log.Fatalf("创建报告文件失败: %v", err)
	}
	defer file.Close()

	// 写入报告头部
	fmt.Fprintf(file, "ZXing Go Wrapper 测试报告\n")
	fmt.Fprintf(file, "生成时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 60))

	// 写入统计信息
	fmt.Fprintf(file, "测试统计:\n")
	fmt.Fprintf(file, "  总图片数: %d\n", report.TotalImages)
	fmt.Fprintf(file, "  成功数: %d\n", report.SuccessCount)
	fmt.Fprintf(file, "  失败数: %d\n", report.FailedCount)
	fmt.Fprintf(file, "  成功率: %.2f%%\n", float64(report.SuccessCount)/float64(report.TotalImages)*100)
	fmt.Fprintf(file, "  总耗时: %v\n", report.TotalTime)
	fmt.Fprintf(file, "  平均耗时: %v\n", report.TotalTime/time.Duration(report.TotalImages))
	fmt.Fprintf(file, "\n")

	// 写入详细结果
	fmt.Fprintf(file, "详细结果:\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("=", 60))
	
	for i, result := range report.Results {
		fmt.Fprintf(file, "\n%d. 图片: %s\n", i+1, filepath.Base(result.ImagePath))
		fmt.Fprintf(file, "   路径: %s\n", result.ImagePath)
		fmt.Fprintf(file, "   状态: %s\n", func() string {
			if result.Success {
				return "成功"
			}
			return "失败"
		}())
		fmt.Fprintf(file, "   解码耗时: %v\n", result.DecodeTime)
		
		if result.Success && result.Result != nil {
			fmt.Fprintf(file, "   单条码结果:\n")
			fmt.Fprintf(file, "     文本: %s\n", result.Result.Text)
			fmt.Fprintf(file, "     格式: %s\n", result.Result.Format.String())
			fmt.Fprintf(file, "     置信度: %.2f\n", result.Result.Confidence)
		}
		
		if len(result.MultiResults) > 0 {
			fmt.Fprintf(file, "   多条码结果 (共%d个):\n", len(result.MultiResults))
			for j, res := range result.MultiResults {
				fmt.Fprintf(file, "     %d. 文本: %s\n", j+1, res.Text)
				fmt.Fprintf(file, "        格式: %s\n", res.Format.String())
				fmt.Fprintf(file, "        置信度: %.2f\n", res.Confidence)
			}
		}
		
		if !result.Success {
			fmt.Fprintf(file, "   错误信息: %s\n", result.Error)
		}
	}

	// 写入性能分析
	fmt.Fprintf(file, "\n性能分析:\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("=", 60))
	
	var totalDecodeTime time.Duration
	var successCount int
	
	for _, result := range report.Results {
		if result.Success {
			totalDecodeTime += result.DecodeTime
			successCount++
		}
	}
	
	if successCount > 0 {
		fmt.Fprintf(file, "平均解码时间: %v\n", totalDecodeTime/time.Duration(successCount))
		fmt.Fprintf(file, "最快解码时间: %v\n", func() time.Duration {
			min := report.Results[0].DecodeTime
			for _, result := range report.Results {
				if result.Success && result.DecodeTime < min {
					min = result.DecodeTime
				}
			}
			return min
		}())
		fmt.Fprintf(file, "最慢解码时间: %v\n", func() time.Duration {
			max := time.Duration(0)
			for _, result := range report.Results {
				if result.Success && result.DecodeTime > max {
					max = result.DecodeTime
				}
			}
			return max
		}())
	}
} 