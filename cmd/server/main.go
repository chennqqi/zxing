package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chennqqi/zxing/pkg/zxing"
	"github.com/gin-gonic/gin"
)

// 上传文件大小限制
const maxUploadSize = 10 * 1024 * 1024 // 10MB

// DecodeResult represents a single barcode decode result.
type DecodeResult struct {
	Text   string `json:"text"`
	Format string `json:"format"`
}

// DecodeResponse represents the API response for decode endpoint.
type DecodeResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Results []*DecodeResult `json:"results,omitempty"`
}

// 解码选项
type DecodeRequest struct {
	Formats      []string `json:"formats"`
	TryHarder    bool     `json:"try_harder"`
	TryRotate    bool     `json:"try_rotate"`
	TryInvert    bool     `json:"try_invert"`
	TryDownscale bool     `json:"try_downscale"`
}

func main() {
	// 创建上传目录
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal(err)
	}

	// 创建 Gin 路由
	r := gin.Default()

	// 设置静态文件目录
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")
	r.LoadHTMLGlob("templates/*")

	// 首页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "ZXing Server",
		})
	})

	// 上传并解码图片
	r.POST("/api/decode", func(c *gin.Context) {
		// 解析请求参数
		var req DecodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, DecodeResponse{
				Success: false,
				Message: "Invalid request parameters",
			})
			return
		}

		// 获取上传的文件
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, DecodeResponse{
				Success: false,
				Message: "No image file uploaded",
			})
			return
		}

		// 检查文件大小
		if file.Size > maxUploadSize {
			c.JSON(http.StatusBadRequest, DecodeResponse{
				Success: false,
				Message: "File too large",
			})
			return
		}

		// 保存文件
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		filepath := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, filepath); err != nil {
			c.JSON(http.StatusInternalServerError, DecodeResponse{
				Success: false,
				Message: "Failed to save file",
			})
			return
		}

		// Open and decode the uploaded image
		f, err := os.Open(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, DecodeResponse{
				Success: false,
				Message: "Failed to open uploaded file",
			})
			return
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			c.JSON(http.StatusBadRequest, DecodeResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to decode image: %v", err),
			})
			return
		}

		// Create ZXing instance
		zx, err := zxing.New(nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, DecodeResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create ZXing instance: %v", err),
			})
			return
		}
		defer zx.Close()

		// Create decode options
		opts := &zxing.DecodeOptions{
			TryHarder: req.TryHarder,
		}
		if len(req.Formats) > 0 {
			opts.PossibleFormats = req.Formats
		}

		// Decode the image
		result, err := zx.DecodeImage(context.Background(), img, opts)
		if err != nil {
			c.JSON(http.StatusOK, DecodeResponse{
				Success: false,
				Message: fmt.Sprintf("Decode failed: %v", err),
			})
			return
		}

		// Convert to API response
		var apiResults []*DecodeResult
		if result != nil && len(result.Text) > 0 {
			apiResults = append(apiResults, &DecodeResult{
				Text:   result.Text,
				Format: result.Format,
			})
		}

		c.JSON(http.StatusOK, DecodeResponse{
			Success: true,
			Results: apiResults,
		})
	})

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
