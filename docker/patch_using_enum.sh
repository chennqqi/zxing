#!/bin/sh
# Patch zxing-cpp to replace 'using enum BarcodeFormat;' with #define macros
# for GCC 10 compatibility (GCC 11+ supports 'using enum' from C++20).
#
# GCC 10 does not support:
#   - using enum BarcodeFormat;  (P0648R2, GCC 11+)
#   - using BarcodeFormat::None; (same feature)
#
# Workaround: use #define macros to alias each enum value.
# These are local to each .cpp file and do not leak into headers.
#
# Usage: ./patch_using_enum.sh /path/to/zxing-cpp

set -e

SRCDIR="${1:-.}"

# All enum values from BarcodeFormat (via ZX_BCF_LIST in BarcodeFormat.h)
# Extract dynamically from the header
ENUM_VALUES=$(grep -E '^\s+X\(' "$SRCDIR/core/src/BarcodeFormat.h" | \
    sed "s/.*X(\([^,]*\),.*/\1/" | tr -d ' ')

if [ -z "$ENUM_VALUES" ]; then
    echo "ERROR: Could not extract enum values from BarcodeFormat.h"
    exit 1
fi

# Build the replacement block
# Note: all 'using enum BarcodeFormat;' statements are inside 'namespace ZXing {}',
# so we use 'BarcodeFormat::X' without 'ZXing::' prefix to avoid double qualification.
# We use 'static constexpr auto' instead of #define because #define would also expand
# in already-qualified names like 'BarcodeFormat::AllGS1' -> 'BarcodeFormat::BarcodeFormat::AllGS1'.
DEFINES=""
for val in $ENUM_VALUES; do
    DEFINES="${DEFINES}static constexpr auto ${val} = BarcodeFormat::${val}; "
done

# Find and patch all files containing 'using enum BarcodeFormat;'
FILES=$(grep -rl 'using enum BarcodeFormat;' "$SRCDIR" --include='*.cpp' --include='*.h' 2>/dev/null || true)

if [ -z "$FILES" ]; then
    echo "No files to patch"
    exit 0
fi

for f in $FILES; do
    echo "Patching: $f"
    # Replace 'using enum BarcodeFormat;' with #define macros
    # Use a temp file for the replacement to handle multi-line
    python3 -c "
import sys
with open('$f', 'r', encoding='utf-8') as fh:
    content = fh.read()
defines = '''$DEFINES'''
content = content.replace('using enum BarcodeFormat;', defines)
with open('$f', 'w', encoding='utf-8') as fh:
    fh.write(content)
"
done

echo "Patch applied successfully to $(echo "$FILES" | wc -w) file(s)"
