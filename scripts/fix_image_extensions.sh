#!/bin/bash

# 脚本：根据 magic number 修正图片扩展名
# 用途：修正 data/images 目录中 SHA256 命名的图片文件扩展名

set -e

IMAGE_DIR="${1:-data/images}"

if [ ! -d "$IMAGE_DIR" ]; then
    echo "Error: Directory $IMAGE_DIR does not exist"
    exit 1
fi

# 图片格式的 magic number 映射
# 格式: "magic_bytes:extension"
declare -A MAGIC_MAP=(
    # PNG: 89 50 4E 47 0D 0A 1A 0A
    ["89504e47"]="png"
    # JPEG: FF D8 FF
    ["ffd8ff"]="jpg"
    # GIF: 47 49 46 38 (GIF8)
    ["47494638"]="gif"
    # WebP: 52 49 46 46 ... 57 45 42 50 (RIFF...WEBP)
    ["52494646"]="webp"
    # BMP: 42 4D (BM)
    ["424d"]="bmp"
    # TIFF: 49 49 2A 00 (little-endian) or 4D 4D 00 2A (big-endian)
    ["49492a00"]="tif"
    ["4d4d002a"]="tif"
)

# 函数：检测文件类型
detect_image_type() {
    local file="$1"
    
    # 读取文件的前 16 字节
    local hex=$(hexdump -n 16 -v -e '/1 "%02x"' "$file" 2>/dev/null)
    
    if [ -z "$hex" ]; then
        return 1
    fi
    
    # 检查 PNG
    if [[ "${hex:0:8}" == "89504e47" ]]; then
        echo "png"
        return 0
    fi
    
    # 检查 JPEG
    if [[ "${hex:0:6}" == "ffd8ff" ]]; then
        echo "jpg"
        return 0
    fi
    
    # 检查 GIF
    if [[ "${hex:0:8}" == "47494638" ]]; then
        echo "gif"
        return 0
    fi
    
    # 检查 WebP (RIFF...WEBP)
    if [[ "${hex:0:8}" == "52494646" ]]; then
        # 读取更多字节来确认 WEBP 标识（在偏移 12-15 应该是 "WEBP"）
        local more_hex=$(hexdump -n 16 -v -e '/1 "%02x"' "$file" 2>/dev/null)
        if [[ "${more_hex:24:8}" == "57454250" ]]; then
            echo "webp"
            return 0
        fi
        # 也尝试使用 file 命令作为后备
        if file -b --mime-type "$file" 2>/dev/null | grep -q "webp"; then
            echo "webp"
            return 0
        fi
    fi
    
    # 检查 BMP
    if [[ "${hex:0:4}" == "424d" ]]; then
        echo "bmp"
        return 0
    fi
    
    # 检查 TIFF
    if [[ "${hex:0:8}" == "49492a00" ]] || [[ "${hex:0:8}" == "4d4d002a" ]]; then
        echo "tif"
        return 0
    fi
    
    # 使用 file 命令作为后备
    local mime=$(file -b --mime-type "$file" 2>/dev/null)
    case "$mime" in
        image/png)
            echo "png"
            return 0
            ;;
        image/jpeg)
            echo "jpg"
            return 0
            ;;
        image/gif)
            echo "gif"
            return 0
            ;;
        image/webp)
            echo "webp"
            return 0
            ;;
        image/bmp)
            echo "bmp"
            return 0
            ;;
        image/tiff)
            echo "tif"
            return 0
            ;;
    esac
    
    return 1
}

# 处理文件
fixed_count=0
skipped_count=0
error_count=0

echo "Scanning directory: $IMAGE_DIR"
echo ""

# 遍历目录中的文件
while IFS= read -r -d '' file; do
    filename=$(basename "$file")
    
    # 跳过已有扩展名的文件
    if [[ "$filename" =~ \.(jpg|jpeg|png|gif|webp|bmp|tif|tiff)$ ]]; then
        skipped_count=$((skipped_count + 1))
        continue
    fi
    
    # 检测文件类型
    ext=$(detect_image_type "$file" || echo "")
    
    if [ -z "$ext" ]; then
        echo "⚠️  Cannot determine type: $filename"
        error_count=$((error_count + 1))
        continue
    fi
    
    # 构建新文件名
    new_filename="${filename}.${ext}"
    new_path="${IMAGE_DIR}/${new_filename}"
    
    # 检查新文件名是否已存在
    if [ -e "$new_path" ] && [ "$file" != "$new_path" ]; then
        echo "⚠️  Target exists, skipping: $filename -> $new_filename"
        error_count=$((error_count + 1))
        continue
    fi
    
    # 重命名文件
    mv "$file" "$new_path"
    echo "✓ Renamed: $filename -> $new_filename"
    fixed_count=$((fixed_count + 1))
    
done < <(find "$IMAGE_DIR" -maxdepth 1 -type f -print0)

echo ""
echo "Summary:"
echo "  Fixed:   $fixed_count"
echo "  Skipped: $skipped_count (already have extension)"
echo "  Errors:  $error_count"
echo ""
echo "Done!"
