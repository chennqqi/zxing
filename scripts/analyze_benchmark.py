#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import re
import sys
import json
from datetime import datetime
from pathlib import Path

def parse_benchmark_output(output):
    """解析性能测试输出"""
    results = {}
    current_benchmark = None
    
    for line in output.split('\n'):
        # 匹配基准测试名称
        benchmark_match = re.match(r'^Benchmark(\w+)\s+(\d+)\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op$', line)
        if benchmark_match:
            name = benchmark_match.group(1)
            iterations = int(benchmark_match.group(2))
            ns_per_op = float(benchmark_match.group(3))
            bytes_per_op = int(benchmark_match.group(4))
            allocs_per_op = int(benchmark_match.group(5))
            
            if name not in results:
                results[name] = []
            
            results[name].append({
                'iterations': iterations,
                'ns_per_op': ns_per_op,
                'bytes_per_op': bytes_per_op,
                'allocs_per_op': allocs_per_op,
                'timestamp': datetime.now().isoformat()
            })
    
    return results

def analyze_results(results):
    """分析性能测试结果"""
    analysis = {}
    
    for name, runs in results.items():
        if not runs:
            continue
            
        # 计算平均值
        ns_per_op_avg = sum(r['ns_per_op'] for r in runs) / len(runs)
        bytes_per_op_avg = sum(r['bytes_per_op'] for r in runs) / len(runs)
        allocs_per_op_avg = sum(r['allocs_per_op'] for r in runs) / len(runs)
        
        # 计算标准差
        ns_per_op_std = (sum((r['ns_per_op'] - ns_per_op_avg) ** 2 for r in runs) / len(runs)) ** 0.5
        bytes_per_op_std = (sum((r['bytes_per_op'] - bytes_per_op_avg) ** 2 for r in runs) / len(runs)) ** 0.5
        allocs_per_op_std = (sum((r['allocs_per_op'] - allocs_per_op_avg) ** 2 for r in runs) / len(runs)) ** 0.5
        
        analysis[name] = {
            'ns_per_op': {
                'avg': ns_per_op_avg,
                'std': ns_per_op_std,
                'unit': 'ns/op'
            },
            'bytes_per_op': {
                'avg': bytes_per_op_avg,
                'std': bytes_per_op_std,
                'unit': 'B/op'
            },
            'allocs_per_op': {
                'avg': allocs_per_op_avg,
                'std': allocs_per_op_std,
                'unit': 'allocs/op'
            }
        }
    
    return analysis

def save_results(results, analysis, output_dir):
    """保存测试结果和分析"""
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # 保存原始结果
    with open(output_dir / 'benchmark_results.json', 'w') as f:
        json.dump(results, f, indent=2)
    
    # 保存分析结果
    with open(output_dir / 'benchmark_analysis.json', 'w') as f:
        json.dump(analysis, f, indent=2)
    
    # 生成 Markdown 报告
    with open(output_dir / 'benchmark_report.md', 'w') as f:
        f.write('# 性能测试报告\n\n')
        f.write(f'生成时间: {datetime.now().isoformat()}\n\n')
        
        for name, metrics in analysis.items():
            f.write(f'## {name}\n\n')
            for metric_name, data in metrics.items():
                f.write(f'### {metric_name}\n\n')
                f.write(f'- 平均值: {data["avg"]:.2f} {data["unit"]}\n')
                f.write(f'- 标准差: {data["std"]:.2f} {data["unit"]}\n\n')

def main():
    if len(sys.argv) != 3:
        print(f'Usage: {sys.argv[0]} <benchmark_output_file> <output_dir>')
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_dir = sys.argv[2]
    
    with open(input_file, 'r') as f:
        output = f.read()
    
    results = parse_benchmark_output(output)
    analysis = analyze_results(results)
    save_results(results, analysis, output_dir)
    
    print(f'分析完成，结果已保存到 {output_dir}')

if __name__ == '__main__':
    main() 