#include "zxing.h"
#include <ZXing/ReadBarcode.h>
#include <memory>
#include <string>
#include <vector>
#include <cstdarg>
#include <cstring>

using namespace ZXing;

// 错误信息缓冲区
static char last_error[256] = {0};

// 设置错误信息
static void set_error(const char* format, ...) {
    va_list args;
    va_start(args, format);
    vsnprintf(last_error, sizeof(last_error), format, args);
    va_end(args);
}

// 创建默认解码选项
DecodeOptions* create_default_options() {
    DecodeOptions* options = new DecodeOptions();
    if (!options) {
        set_error("Failed to allocate memory for options");
        return nullptr;
    }
    
    options->formats = FORMAT_ALL;
    options->try_harder = 1;
    options->try_rotate = 1;
    options->try_invert = 0;
    options->try_downscale = 1;
    
    return options;
}

// 释放解码选项
void free_options(DecodeOptions* options) {
    delete options;
}

// 转换 ZXing 格式到 C 格式
static ::BarcodeFormat convert_format(ZXing::BarcodeFormat format) {
    switch (format) {
        case ZXing::BarcodeFormat::QR_CODE: return FORMAT_QR_CODE;
        case ZXing::BarcodeFormat::AZTEC: return FORMAT_AZTEC;
        case ZXing::BarcodeFormat::CODABAR: return FORMAT_CODABAR;
        case ZXing::BarcodeFormat::CODE_39: return FORMAT_CODE_39;
        case ZXing::BarcodeFormat::CODE_93: return FORMAT_CODE_93;
        case ZXing::BarcodeFormat::CODE_128: return FORMAT_CODE_128;
        case ZXing::BarcodeFormat::DATA_MATRIX: return FORMAT_DATA_MATRIX;
        case ZXing::BarcodeFormat::EAN_8: return FORMAT_EAN_8;
        case ZXing::BarcodeFormat::EAN_13: return FORMAT_EAN_13;
        case ZXing::BarcodeFormat::ITF: return FORMAT_ITF;
        case ZXing::BarcodeFormat::MAXICODE: return FORMAT_MAXICODE;
        case ZXing::BarcodeFormat::PDF_417: return FORMAT_PDF_417;
        case ZXing::BarcodeFormat::UPC_A: return FORMAT_UPC_A;
        case ZXing::BarcodeFormat::UPC_E: return FORMAT_UPC_E;
        default: return FORMAT_NONE;
    }
}

// 解码单个条码
DecodeResult* decode_barcode(const char* image_path, const DecodeOptions* options) {
    try {
        // 加载图像
        auto image = ImageView::FromFile(image_path);
        if (!image) {
            set_error("Failed to load image: %s", image_path);
            return nullptr;
        }

        // 设置解码选项
        DecodeHints hints;
        hints.setTryHarder(options->try_harder);
        hints.setTryRotate(options->try_rotate);
        hints.setTryInvert(options->try_invert);
        hints.setTryDownscale(options->try_downscale);

        // 解码
        auto results = ReadBarcodes(image, hints);
        if (results.empty()) {
            set_error("No barcode found");
            return nullptr;
        }

        // 创建结果
        DecodeResult* result = new DecodeResult();
        if (!result) {
            set_error("Failed to allocate memory for result");
            return nullptr;
        }

        // 填充结果
        result->text = strdup(results[0].text().c_str());
        result->format = convert_format(results[0].format());
        result->confidence = results[0].confidence();

        return result;
    } catch (const std::exception& e) {
        set_error("Decode error: %s", e.what());
        return nullptr;
    }
}

// 解码多个条码
DecodeResult** decode_barcodes(const char* image_path, const DecodeOptions* options, int* count) {
    try {
        // 加载图像
        auto image = ImageView::FromFile(image_path);
        if (!image) {
            set_error("Failed to load image: %s", image_path);
            return nullptr;
        }

        // 设置解码选项
        DecodeHints hints;
        hints.setTryHarder(options->try_harder);
        hints.setTryRotate(options->try_rotate);
        hints.setTryInvert(options->try_invert);
        hints.setTryDownscale(options->try_downscale);

        // 解码
        auto results = ReadBarcodes(image, hints);
        if (results.empty()) {
            set_error("No barcode found");
            return nullptr;
        }

        // 创建结果数组
        *count = results.size();
        DecodeResult** decode_results = new DecodeResult*[*count];
        if (!decode_results) {
            set_error("Failed to allocate memory for results");
            return nullptr;
        }

        // 填充结果
        for (int i = 0; i < *count; i++) {
            decode_results[i] = new DecodeResult();
            if (!decode_results[i]) {
                set_error("Failed to allocate memory for result %d", i);
                for (int j = 0; j < i; j++) {
                    free_result(decode_results[j]);
                }
                delete[] decode_results;
                return nullptr;
            }

            decode_results[i]->text = strdup(results[i].text().c_str());
            decode_results[i]->format = convert_format(results[i].format());
            decode_results[i]->confidence = results[i].confidence();
        }

        return decode_results;
    } catch (const std::exception& e) {
        set_error("Decode error: %s", e.what());
        return nullptr;
    }
}

// 获取解码结果
DecodeResult* decode_result_get(DecodeResult** results, int index) {
    if (!results || index < 0) {
        return nullptr;
    }
    return results[index];
}

// 释放单个解码结果
void free_result(DecodeResult* result) {
    if (result) {
        free(result->text);
        delete result;
    }
}

// 释放多个解码结果
void free_results(DecodeResult** results, int count) {
    if (results) {
        for (int i = 0; i < count; i++) {
            free_result(results[i]);
        }
        delete[] results;
    }
}

// 获取错误信息
const char* get_last_error() {
    return last_error;
} 