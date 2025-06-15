#!/bin/bash

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

# 创建临时构建目录
BUILD_DIR=$(mktemp -d)
trap 'rm -rf "$BUILD_DIR"' EXIT

# 创建 debian 目录结构
mkdir -p "$BUILD_DIR/debian"
mkdir -p "$BUILD_DIR/usr/local/bin"
mkdir -p "$BUILD_DIR/usr/local/lib"
mkdir -p "$BUILD_DIR/usr/local/include"

# 复制文件
cp build/libzxing.so "$BUILD_DIR/usr/local/lib/"
cp include/zxing.h "$BUILD_DIR/usr/local/include/"
cp build/zxing.dll "$BUILD_DIR/usr/local/bin/" 2>/dev/null || true

# 创建 control 文件
cat > "$BUILD_DIR/debian/control" << EOF
Package: zxing
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Depends: libc6 (>= 2.17), libstdc++6 (>= 4.8.1)
Maintainer: Your Name <your.email@example.com>
Description: ZXing barcode scanning library
 ZXing is a barcode scanning library that supports multiple formats
 including QR Code, Data Matrix, UPC, EAN, Code 39, Code 128, etc.
EOF

# 创建 rules 文件
cat > "$BUILD_DIR/debian/rules" << EOF
#!/usr/bin/make -f

%:
	dh \$@

override_dh_auto_install:
	dh_auto_install
	install -D -m 644 usr/local/lib/libzxing.so debian/zxing/usr/local/lib/
	install -D -m 644 usr/local/include/zxing.h debian/zxing/usr/local/include/
	[ -f usr/local/bin/zxing.dll ] && install -D -m 644 usr/local/bin/zxing.dll debian/zxing/usr/local/bin/ || true
EOF

chmod +x "$BUILD_DIR/debian/rules"

# 创建 changelog 文件
cat > "$BUILD_DIR/debian/changelog" << EOF
zxing ($VERSION) stable; urgency=medium

  * Initial release

 -- Your Name <your.email@example.com>  $(date -R)
EOF

# 创建 compat 文件
echo "9" > "$BUILD_DIR/debian/compat"

# 创建 source/format 文件
mkdir -p "$BUILD_DIR/debian/source"
echo "3.0 (native)" > "$BUILD_DIR/debian/source/format"

# 构建包
cd "$BUILD_DIR"
dpkg-buildpackage -us -uc -b

# 移动生成的包到 dist 目录
mkdir -p ../../dist
mv ../zxing_${VERSION}_amd64.deb ../../dist/ 