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

# 创建 RPM 构建目录结构
mkdir -p "$BUILD_DIR/BUILD"
mkdir -p "$BUILD_DIR/RPMS"
mkdir -p "$BUILD_DIR/SOURCES"
mkdir -p "$BUILD_DIR/SPECS"
mkdir -p "$BUILD_DIR/SRPMS"

# 创建 spec 文件
cat > "$BUILD_DIR/SPECS/zxing.spec" << EOF
Name:           zxing
Version:        $VERSION
Release:        1%{?dist}
Summary:        ZXing barcode scanning library

License:        MIT
URL:            https://github.com/yourusername/zxing
BuildArch:      x86_64

Requires:       libc.so.6()(64bit), libstdc++.so.6()(64bit)

%description
ZXing is a barcode scanning library that supports multiple formats
including QR Code, Data Matrix, UPC, EAN, Code 39, Code 128, etc.

%prep
# Nothing to do here

%build
# Nothing to do here

%install
mkdir -p %{buildroot}/usr/local/lib
mkdir -p %{buildroot}/usr/local/include
mkdir -p %{buildroot}/usr/local/bin

install -m 644 build/libzxing.so %{buildroot}/usr/local/lib/
install -m 644 include/zxing.h %{buildroot}/usr/local/include/
[ -f build/zxing.dll ] && install -m 644 build/zxing.dll %{buildroot}/usr/local/bin/ || true

%files
/usr/local/lib/libzxing.so
/usr/local/include/zxing.h
/usr/local/bin/zxing.dll

%changelog
* $(date "+%a %b %d %Y") Your Name <your.email@example.com> - $VERSION-1
- Initial release
EOF

# 构建 RPM 包
rpmbuild --define "_topdir $BUILD_DIR" -bb "$BUILD_DIR/SPECS/zxing.spec"

# 移动生成的包到 dist 目录
mkdir -p ../../dist
mv "$BUILD_DIR/RPMS/x86_64/zxing-${VERSION}-1.x86_64.rpm" ../../dist/ 