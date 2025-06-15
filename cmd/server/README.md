# ZXing Server

ZXing Server 是一个基于 ZXing 的条码解码服务，提供 Web 界面和 REST API。

## 功能特点

- 支持多种条码格式：
  - QR Code
  - Aztec
  - Codabar
  - Code 39
  - Code 93
  - Code 128
  - Data Matrix
  - EAN-8
  - EAN-13
  - ITF
  - MaxiCode
  - PDF417
  - UPC-A
  - UPC-E

- 高级解码选项：
  - 尝试更努力的解码
  - 尝试旋转图像
  - 尝试反转图像
  - 尝试缩小图像

- 用户友好的 Web 界面
- RESTful API
- Docker 支持

## 快速开始

### 使用 Docker

1. 克隆仓库：
```bash
git clone https://github.com/threatbook/zxing-server.git
cd zxing-server
```

2. 启动服务：
```bash
docker-compose up -d
```

3. 访问 Web 界面：
```
http://localhost:8080
```

### 手动构建

1. 安装依赖：
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y build-essential cmake libzxing-dev

# CentOS/RHEL
sudo yum install -y gcc gcc-c++ make cmake zxing-cpp-devel

# macOS
brew install cmake zxing-cpp
```

2. 构建服务：
```bash
go build -o zxing-server
```

3. 运行服务：
```bash
./zxing-server
```

## API 文档

### 解码图片

```
POST /api/decode
```

请求参数：
- `image`: 图片文件（multipart/form-data）
- `options`: JSON 格式的解码选项
  ```json
  {
    "formats": ["QR_CODE", "EAN_13"],
    "try_harder": true,
    "try_rotate": true,
    "try_invert": true,
    "try_downscale": true
  }
  ```

响应：
```json
{
  "success": true,
  "results": [
    {
      "text": "解码文本",
      "format": "QR Code",
      "confidence": 0.95
    }
  ]
}
```

## 开发

### 项目结构

```
.
├── Dockerfile          # Docker 构建文件
├── docker-compose.yml  # Docker Compose 配置
├── go.mod             # Go 模块文件
├── main.go            # 主程序
├── templates/         # HTML 模板
│   └── index.html     # 主页模板
├── static/           # 静态文件
└── uploads/          # 上传文件目录
```

### 构建测试

```bash
go test ./...
```

## 许可证

MIT License 