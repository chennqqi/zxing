package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chennqqi/zxing"
	"github.com/gin-gonic/gin"
)

// 上传文件大小限制
const maxUploadSize = 10 * 1024 * 1024 // 10MB

// 解码结果
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

		// 创建解码选项
		opts := zxing.NewDefaultOptions()
		if opts == nil {
			c.JSON(http.StatusInternalServerError, DecodeResponse{
				Success: false,
				Message: "Failed to create decode options",
			})
			return
		}

		// 设置解码选项
		opts.TryHarder = req.TryHarder
		opts.TryRotate = req.TryRotate
		opts.TryInvert = req.TryInvert
		opts.TryDownscale = req.TryDownscale

		// 设置解码格式
		if len(req.Formats) > 0 {
			opts.Formats = 0
			for _, format := range req.Formats {
				switch format {
				case "QR_CODE":
					opts.Formats |= zxing.FormatQRCode
				case "AZTEC":
					opts.Formats |= zxing.FormatAztec
				case "CODABAR":
					opts.Formats |= zxing.FormatCodabar
				case "CODE_39":
					opts.Formats |= zxing.FormatCode39
				case "CODE_93":
					opts.Formats |= zxing.FormatCode93
				case "CODE_128":
					opts.Formats |= zxing.FormatCode128
				case "DATA_MATRIX":
					opts.Formats |= zxing.FormatDataMatrix
				case "EAN_8":
					opts.Formats |= zxing.FormatEAN8
				case "EAN_13":
					opts.Formats |= zxing.FormatEAN13
				case "ITF":
					opts.Formats |= zxing.FormatITF
				case "MAXICODE":
					opts.Formats |= zxing.FormatMaxiCode
				case "PDF_417":
					opts.Formats |= zxing.FormatPDF417
				case "UPC_A":
					opts.Formats |= zxing.FormatUPCA
				case "UPC_E":
					opts.Formats |= zxing.FormatUPCE
				}
			}
		}

		// 解码图片
		results, err := zxing.DecodeMulti(filepath, opts)
		if err != nil {
			c.JSON(http.StatusOK, DecodeResponse{
				Success: false,
				Message: fmt.Sprintf("Decode failed: %v", err),
			})
			return
		}

		// 返回结果
		c.JSON(http.StatusOK, DecodeResponse{
			Success: true,
			Results: results,
		})
	})

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
