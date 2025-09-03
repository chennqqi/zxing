# ZXing Go WASM 集成 Makefile

.PHONY: all build test clean wasm cgo example help

# 默认目标
all: build

# 构建所有目标
build: build-go build-example

# 构建 Go 包
build-go:
	@echo "构建 Go 包..."
	go build ./pkg/zxing/
	go build ./pkg/wasm/

# 构建示例程序
build-example:
	@echo "构建示例程序..."
	mkdir -p bin
	go build -o bin/wasm-example ./cmd/wasm-example/

# 构建 WASM 版本
build-wasm:
	@echo "构建 WASM 版本..."
	mkdir -p wasm
	GOOS=js GOARCH=wasm go build -o wasm/wasm-example.wasm ./cmd/wasm-example/
	@if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm/; \
		echo "已复制 wasm_exec.js"; \
	fi

# 构建 CGO 版本
build-cgo:
	@echo "构建 CGO 版本..."
	mkdir -p bin
	CGO_ENABLED=1 go build -tags cgo -o bin/zxing-cgo ./cmd/wasm-example/

# 运行测试
test:
	@echo "运行测试..."
	go test ./pkg/zxing/ -v
	go test ./pkg/wasm/ -v

# 运行基准测试
bench:
	@echo "运行基准测试..."
	go test -bench=. ./pkg/zxing/

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -rf wasm/

# 运行示例
run-example:
	@echo "运行示例程序..."
	go run ./cmd/wasm-example/

# 显示帮助
help:
	@echo "可用的构建目标:"
	@echo "  all          - 构建所有目标"
	@echo "  build        - 构建 Go 包和示例"
	@echo "  build-go     - 构建 Go 包"
	@echo "  build-example- 构建示例程序"
	@echo "  build-wasm   - 构建 WASM 版本"
	@echo "  build-cgo    - 构建 CGO 版本"
	@echo "  test         - 运行测试"
	@echo "  bench        - 运行基准测试"
	@echo "  run-example  - 运行示例程序"
	@echo "  clean        - 清理构建文件"
	@echo "  help         - 显示此帮助信息"