#!/bin/bash

# 设置测试次数和输出目录
BENCHMARK_COUNT=5
OUTPUT_DIR="benchmark_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="${OUTPUT_DIR}/benchmark_${TIMESTAMP}.txt"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 运行所有性能测试并保存结果
echo "Running all benchmarks..."
go test -bench=. -benchmem -count=$BENCHMARK_COUNT ./... | tee "$OUTPUT_FILE"

# 分析结果
echo -e "\nAnalyzing benchmark results..."
python3 scripts/analyze_benchmark.py "$OUTPUT_FILE" "${OUTPUT_DIR}/${TIMESTAMP}"

# 显示报告
echo -e "\nBenchmark report:"
cat "${OUTPUT_DIR}/${TIMESTAMP}/benchmark_report.md"

# 运行单个条码解码性能测试
echo -e "\nRunning single barcode decode benchmark..."
go test -bench=BenchmarkDecode -benchmem -count=$BENCHMARK_COUNT ./...

# 运行多个条码解码性能测试
echo -e "\nRunning multiple barcode decode benchmark..."
go test -bench=BenchmarkDecodeMulti -benchmem -count=$BENCHMARK_COUNT ./...

# 运行带高级选项的单条码解码性能测试
echo -e "\nRunning single barcode decode with options benchmark..."
go test -bench=BenchmarkDecodeWithOptions -benchmem -count=$BENCHMARK_COUNT ./...

# 运行带高级选项的多条码解码性能测试
echo -e "\nRunning multiple barcode decode with options benchmark..."
go test -bench=BenchmarkDecodeMultiWithOptions -benchmem -count=$BENCHMARK_COUNT ./... 