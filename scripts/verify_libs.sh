#!/bin/bash
# Verify static library files in the lib/ directory.
#
# Checks each platform directory lib/{os}-{arch}/ for libZXing.a and
# libzxingwrapper.a. Missing platform directories are reported but do not
# fail the check — a directory that EXISTS but has missing/corrupt files
# does fail.

set -e

echo "========================================"
echo "Verifying Library Files"
echo "========================================"
echo ""

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

ALL_FOUND=true
FOUND_ANY=false

# Check a single platform directory
check_platform() {
    local name="$1"
    local dir="$PROJECT_ROOT/lib/$name"

    echo "Checking $name..."

    if [ ! -d "$dir" ]; then
        echo "  [SKIP] Directory not present: lib/$name"
        echo ""
        return
    fi

    FOUND_ANY=true
    for lib in libZXing.a libzxingwrapper.a; do
        if [ -f "$dir/$lib" ]; then
            SIZE=$(du -h "$dir/$lib" | cut -f1)
            echo "  [OK] $lib: $SIZE"
        else
            echo "  [MISSING] $lib"
            ALL_FOUND=false
        fi
    done
    echo ""
}

check_platform "linux-x64"
check_platform "linux-arm64"
check_platform "windows-x64"
check_platform "windows-arm64"
check_platform "darwin-x64"
check_platform "darwin-arm64"

# Check WASM files
echo "Checking WASM files..."
WASM_DIR="$PROJECT_ROOT/wasm"

if [ -d "$WASM_DIR" ]; then
    if [ -f "$WASM_DIR/zxingwrapper.wasm" ]; then
        SIZE=$(du -h "$WASM_DIR/zxingwrapper.wasm" | cut -f1)
        echo "  [OK] zxingwrapper.wasm: $SIZE"
    else
        echo "  [MISSING] zxingwrapper.wasm"
        ALL_FOUND=false
    fi
else
    echo "  [MISSING] Directory: $WASM_DIR"
    ALL_FOUND=false
fi

echo ""
echo "========================================"
if [ "$ALL_FOUND" = true ]; then
    echo "All present library files verified!"
    if [ "$FOUND_ANY" = false ]; then
        echo "Note: no platform lib/ directories found."
    fi
    exit 0
else
    echo "Some library files are missing!"
    echo "Please run the build scripts to generate them."
    exit 1
fi
