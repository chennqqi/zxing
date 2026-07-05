#!/usr/bin/env python3
"""
修正data/images目录中SHA256命名文件的扩展名
根据文件的magic number（文件头）判断实际格式并修正扩展名
"""

import os
import sys
from pathlib import Path

# 常见图片格式的magic number
MAGIC_NUMBERS = {
    b'\x89PNG\r\n\x1a\n': '.png',
    b'\xff\xd8\xff': '.jpg',  # JPEG
    b'GIF87a': '.gif',
    b'GIF89a': '.gif',
    b'RIFF': '.webp',  # 需要进一步检查
    b'BM': '.bmp',
    b'II*\x00': '.tif',  # TIFF little-endian
    b'MM\x00*': '.tif',  # TIFF big-endian
}

def detect_image_format(file_path):
    """根据文件头检测图片格式"""
    try:
        with open(file_path, 'rb') as f:
            header = f.read(32)  # 读取更多字节以检测压缩格式
            
        # 检查PNG
        if header.startswith(b'\x89PNG\r\n\x1a\n'):
            return '.png'
        
        # 检查JPEG
        if header.startswith(b'\xff\xd8\xff'):
            return '.jpg'
        
        # 检查GIF
        if header.startswith(b'GIF87a') or header.startswith(b'GIF89a'):
            return '.gif'
        
        # 检查WebP (RIFF...WEBP)
        if header.startswith(b'RIFF') and b'WEBP' in header:
            return '.webp'
        
        # 检查BMP
        if header.startswith(b'BM'):
            return '.bmp'
        
        # 检查TIFF
        if header.startswith(b'II*\x00') or header.startswith(b'MM\x00*'):
            return '.tif'
        
        # 检查PNG压缩格式（zlib压缩，以78 9c开头）
        # 某些PNG可能被压缩，尝试检测zlib magic number
        if header[0:2] == b'\x78\x9c' or header[0:2] == b'\x78\x01' or header[0:2] == b'\x78\xda':
            # 进一步检查：可能是PNG的IDAT块
            # 读取更多数据来确认
            with open(file_path, 'rb') as f2:
                full_data = f2.read(1024)
                # 如果包含PNG特征，可能是PNG
                if b'PNG' in full_data or b'IHDR' in full_data:
                    return '.png'
                # 否则可能是其他zlib压缩格式，但根据上下文，很可能是PNG
                return '.png'
        
        return None
    except Exception as e:
        print(f"Error reading {file_path}: {e}", file=sys.stderr)
        return None

def fix_image_extensions(image_dir):
    """修正图片目录中的文件扩展名"""
    image_path = Path(image_dir)
    if not image_path.exists():
        print(f"Error: Directory {image_dir} does not exist", file=sys.stderr)
        return
    
    fixed_count = 0
    skipped_count = 0
    error_count = 0
    
    for file_path in image_path.iterdir():
        if not file_path.is_file():
            continue
        
        # 跳过已经有扩展名的文件
        if file_path.suffix:
            skipped_count += 1
            continue
        
        # 检测文件格式
        detected_ext = detect_image_format(file_path)
        
        if detected_ext:
            new_path = file_path.with_suffix(detected_ext)
            
            # 如果目标文件已存在，跳过
            if new_path.exists():
                print(f"Skipping {file_path.name}: target {new_path.name} already exists")
                skipped_count += 1
                continue
            
            try:
                file_path.rename(new_path)
                print(f"Renamed: {file_path.name} -> {new_path.name}")
                fixed_count += 1
            except Exception as e:
                print(f"Error renaming {file_path.name}: {e}", file=sys.stderr)
                error_count += 1
        else:
            print(f"Unknown format: {file_path.name}")
            error_count += 1
    
    print(f"\nSummary:")
    print(f"  Fixed: {fixed_count}")
    print(f"  Skipped: {skipped_count}")
    print(f"  Errors: {error_count}")

if __name__ == '__main__':
    if len(sys.argv) > 1:
        image_dir = sys.argv[1]
    else:
        # 默认使用项目根目录下的data/images
        script_dir = Path(__file__).parent
        image_dir = script_dir.parent / 'data' / 'images'
    
    fix_image_extensions(image_dir)
