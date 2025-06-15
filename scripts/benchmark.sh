#!/bin/bash

# 设置测试次数
BENCHMARK_COUNT=5

# 运行所有性能测试
echo "Running all benchmarks..."
go test -bench=. -benchmem -count=$BENCHMARK_COUNT ./...

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