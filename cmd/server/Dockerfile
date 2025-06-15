# 构建阶段
FROM golang:1.19 AS builder

# 安装依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    libzxing-dev

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o zxing-server

# 运行阶段
FROM ubuntu:20.04

# 安装运行时依赖
RUN apt-get update && apt-get install -y \
    libzxing-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/zxing-server .

# 复制模板和静态文件
COPY templates ./templates
COPY static ./static

# 创建上传目录
RUN mkdir -p uploads

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./zxing-server"] 