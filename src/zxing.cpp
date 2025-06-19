#include "zxing.h"
#include "zxing_internal.h"
#include <ZXing/ReadBarcode.h>
#include <memory>
#include <string>
#include <vector>
#include <cstdarg>
#include <cstring>
#include <cstdlib>

// 包含stb_image用于图像加载
#define STB_IMAGE_IMPLEMENTATION
#include <stb_image.h>

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

// 初始化标志
static bool initialized = false;

// 初始化 zxing-cpp
int zxing_init() {
    if (initialized) {
        return 0;
    }
    // ZXing C++库通常不需要显式初始化
    initialized = true;
    return 0;
}

// 清理 zxing-cpp
void zxing_cleanup() {
    // ZXing C++库通常不需要显式清理
    initialized = false;
}

// 加载图像 - 简化实现，直接返回nullptr，因为Go端会直接传文件路径
Image* zxing_load_image(const char* path) {
    // 这个函数在Go wrapper中不会被使用
    // Go端直接传文件路径给decode_barcode函数
    return nullptr;
}

// 释放图像
void zxing_free_image(Image* image) {
    // 简化实现
    if (image) {
        delete image;
    }
}

// 转换 ZXing 格式到 C 格式
static ::BarcodeFormat convert_format(ZXing::BarcodeFormat format) {
    switch (format) {
        case ZXing::BarcodeFormat::QRCode: return FORMAT_QR_CODE;
        case ZXing::BarcodeFormat::Aztec: return FORMAT_AZTEC;
        case ZXing::BarcodeFormat::Codabar: return FORMAT_CODABAR;
        case ZXing::BarcodeFormat::Code39: return FORMAT_CODE_39;
        case ZXing::BarcodeFormat::Code93: return FORMAT_CODE_93;
        case ZXing::BarcodeFormat::Code128: return FORMAT_CODE_128;
        case ZXing::BarcodeFormat::DataMatrix: return FORMAT_DATA_MATRIX;
        case ZXing::BarcodeFormat::EAN8: return FORMAT_EAN_8;
        case ZXing::BarcodeFormat::EAN13: return FORMAT_EAN_13;
        case ZXing::BarcodeFormat::ITF: return FORMAT_ITF;
        case ZXing::BarcodeFormat::MaxiCode: return FORMAT_MAXICODE;
        case ZXing::BarcodeFormat::PDF417: return FORMAT_PDF_417;
        case ZXing::BarcodeFormat::UPCA: return FORMAT_UPC_A;
        case ZXing::BarcodeFormat::UPCE: return FORMAT_UPC_E;
        default: return FORMAT_NONE;
    }
}

// 解码单个条码 - 简化实现，直接使用文件路径
DecodeResultInternal* zxing_decode(const Image* image, int formats, int try_harder, int try_rotate, int try_invert, int try_downscale) {
    // 这个函数在Go wrapper中不会被使用
    // Go端直接调用decode_barcode函数
    return nullptr;
}

// 解码多个条码 - 简化实现
DecodeResultInternal** zxing_decode_multi(const Image* image, int formats, int try_harder, int try_rotate, int try_invert, int try_downscale, int* count) {
    // 这个函数在Go wrapper中不会被使用
    return nullptr;
}

// 释放解码结果
void zxing_free_result(DecodeResultInternal* result) {
    if (result) {
        free(result->text);
        delete result;
    }
}

// 释放多个解码结果
void zxing_free_results(DecodeResultInternal** results, int count) {
    if (results) {
        for (int i = 0; i < count; i++) {
            zxing_free_result(results[i]);
        }
        delete[] results;
    }
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

// 解码单个条码 - 主要实现，直接使用ZXing-cpp
DecodeResult* decode_barcode(const char* image_path, const DecodeOptions* options) {
    try {
        // 使用stb_image加载图像
        int width, height, channels;
        std::unique_ptr<stbi_uc, void (*)(void*)> buffer(
            stbi_load(image_path, &width, &height, &channels, 0),
            stbi_image_free);
        
        if (!buffer) {
            set_error("Failed to load image: %s (%s)", image_path, stbi_failure_reason());
            return nullptr;
        }

        // 创建ImageView
        auto ImageFormatFromChannels = std::array{ImageFormat::None, ImageFormat::Lum, ImageFormat::LumA, ImageFormat::RGB, ImageFormat::RGBA};
        ImageView image{buffer.get(), width, height, ImageFormatFromChannels.at(channels)};

        // 设置解码选项
        ReaderOptions hints;
        hints.setTryHarder(options->try_harder);
        hints.setTryRotate(options->try_rotate);
        hints.setTryInvert(options->try_invert);
        hints.setTryDownscale(options->try_downscale);

        // 解码
        auto result = ReadBarcode(image, hints);
        if (!result.isValid()) {
            set_error("No barcode found");
            return nullptr;
        }

        // 创建结果
        DecodeResult* decode_result = new DecodeResult();
        if (!decode_result) {
            set_error("Failed to allocate memory for result");
            return nullptr;
        }

        // 填充结果
        decode_result->text = strdup(result.text().c_str());
        decode_result->format = convert_format(result.format());
        decode_result->confidence = 1.0f; // 默认置信度为1.0

        return decode_result;
    } catch (const std::exception& e) {
        set_error("Decode error: %s", e.what());
        return nullptr;
    }
}

// 解码多个条码 - 主要实现，直接使用ZXing-cpp
DecodeResult** decode_barcodes(const char* image_path, const DecodeOptions* options, int* count) {
    try {
        // 使用stb_image加载图像
        int width, height, channels;
        std::unique_ptr<stbi_uc, void (*)(void*)> buffer(
            stbi_load(image_path, &width, &height, &channels, 0),
            stbi_image_free);
        
        if (!buffer) {
            set_error("Failed to load image: %s (%s)", image_path, stbi_failure_reason());
            return nullptr;
        }

        // 创建ImageView
        auto ImageFormatFromChannels = std::array{ImageFormat::None, ImageFormat::Lum, ImageFormat::LumA, ImageFormat::RGB, ImageFormat::RGBA};
        ImageView image{buffer.get(), width, height, ImageFormatFromChannels.at(channels)};

        // 设置解码选项
        ReaderOptions hints;
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
            decode_results[i]->confidence = 1.0f; // 默认置信度为1.0
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