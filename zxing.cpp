#include "zxing.h"
#include <ZXing.h>
#include <string>
#include <memory>

using namespace ZXing;

char* decode_qrcode(const char* image_path) {
    try {
        // 读取图片
        auto image = ImageViewFromFile(image_path);
        if (!image) {
            return strdup("Error: Failed to load image");
        }

        // 配置解码选项
        DecodeHints hints;
        hints.setTryHarder(true);
        hints.setTryRotate(true);
        hints.setTryInvert(true);
        hints.setFormats(BarcodeFormat::QR_CODE);

        // 解码
        auto results = ReadBarcodes(*image, hints);
        
        if (results.empty()) {
            return strdup("No QR code found");
        }

        // 返回第一个结果
        return strdup(results[0].text().c_str());
    } catch (const std::exception& e) {
        return strdup(e.what());
    }
}

void free_string(char* str) {
    free(str);
} 