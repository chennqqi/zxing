@echo off

REM 设置测试次数和输出目录
set BENCHMARK_COUNT=5
set OUTPUT_DIR=benchmark_results
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set TIMESTAMP=%datetime:~0,8%_%datetime:~8,6%
set OUTPUT_FILE=%OUTPUT_DIR%\benchmark_%TIMESTAMP%.txt

REM 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM 运行所有性能测试并保存结果
echo Running all benchmarks...
go test -bench=. -benchmem -count=%BENCHMARK_COUNT% ./... > "%OUTPUT_FILE%"

REM 分析结果
echo.
echo Analyzing benchmark results...
python scripts\analyze_benchmark.py "%OUTPUT_FILE%" "%OUTPUT_DIR%\%TIMESTAMP%"

REM 显示报告
echo.
echo Benchmark report:
type "%OUTPUT_DIR%\%TIMESTAMP%\benchmark_report.md" 