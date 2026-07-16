#include "zxing.h"
#include "zxing_internal.h"
#include "ReadBarcode.h"
#include <memory>
#include <string>
#include <vector>
#include <cstdarg>
#include <cstring>
#include <cstdlib>

#ifdef __EMSCRIPTEN__
#include <emscripten.h>
#define EXPORT EMSCRIPTEN_KEEPALIVE
#else
#define EXPORT
#endif

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
        case ZXing::BarcodeFormat::DataBar: return FORMAT_CODABAR; // Map DataBar to Codabar (no direct C equivalent)
        case ZXing::BarcodeFormat::DataBarExp: return FORMAT_CODABAR; // Map DataBarExp to Codabar
        case ZXing::BarcodeFormat::DataBarLtd: return FORMAT_CODABAR; // Map DataBarLtd to Codabar
        case ZXing::BarcodeFormat::MicroQRCode: return FORMAT_QR_CODE; // 使用QRCode作为替代
        case ZXing::BarcodeFormat::RMQRCode: return FORMAT_QR_CODE; // 使用QRCode作为替代
        case ZXing::BarcodeFormat::DXFilmEdge: return FORMAT_CODE_128; // 使用Code128作为替代
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
EXPORT DecodeOptions* create_default_options() {
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

// Configures all fields of an existing decode options structure.
EXPORT void configure_decode_options(DecodeOptions* options, int formats, int try_harder,
                                     int try_rotate, int try_invert, int try_downscale) {
    if (!options) {
        return;
    }
    options->formats = formats;
    options->try_harder = try_harder;
    options->try_rotate = try_rotate;
    options->try_invert = try_invert;
    options->try_downscale = try_downscale;
}

// 释放解码选项
EXPORT void free_options(DecodeOptions* options) {
    delete options;
}

// 解码单个条码 - 主要实现，直接使用ZXing-cpp
EXPORT DecodeResult* decode_barcode(const char* image_path, const DecodeOptions* options) {
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
        hints.setTryHarder(options->try_harder != 0);
        hints.setTryRotate(options->try_rotate != 0);
        hints.setTryInvert(options->try_invert != 0);
        hints.setTryDownscale(options->try_downscale != 0);

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
EXPORT DecodeResult** decode_barcodes(const char* image_path, const DecodeOptions* options, int* count) {
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
        hints.setTryHarder(options->try_harder != 0);
        hints.setTryRotate(options->try_rotate != 0);
        hints.setTryInvert(options->try_invert != 0);
        hints.setTryDownscale(options->try_downscale != 0);

        // 解码
        auto barcodes = ReadBarcodes(image, hints);
        if (barcodes.empty()) {
            set_error("No barcode found");
            return nullptr;
        }

        // 创建结果数组
        *count = static_cast<int>(barcodes.size());
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

            decode_results[i]->text = strdup(barcodes[i].text().c_str());
            decode_results[i]->format = convert_format(barcodes[i].format());
            decode_results[i]->confidence = 1.0f; // 默认置信度为1.0
        }

        return decode_results;
    } catch (const std::exception& e) {
        set_error("Decode error: %s", e.what());
        return nullptr;
    }
}

// 获取解码结果
EXPORT DecodeResult* decode_result_get(DecodeResult** results, int index) {
    if (!results || index < 0) {
        return nullptr;
    }
    return results[index];
}

// 释放单个解码结果
EXPORT void free_result(DecodeResult* result) {
    if (result) {
        free(result->text);
        delete result;
    }
}

// 释放多个解码结果
EXPORT void free_results(DecodeResult** results, int count) {
    if (results) {
        for (int i = 0; i < count; i++) {
            free_result(results[i]);
        }
        delete[] results;
    }
}

// 获取错误信息
EXPORT const char* get_last_error() {
    return last_error;
}

// Decode barcode from raw image file data (PNG/JPEG/BMP etc.)
// Used by wazero runtime which cannot access filesystem
EXPORT DecodeResult* decode_barcode_data(const unsigned char* file_data, int file_size, const DecodeOptions* options) {
    if (!file_data || file_size <= 0) {
        set_error("Invalid image data");
        return nullptr;
    }

    // Load image from memory using stb_image
    int width, height, channels;
    unsigned char* img_data = stbi_load_from_memory(file_data, file_size, &width, &height, &channels, 1);
    if (!img_data) {
        set_error("Failed to load image from data: %s", stbi_failure_reason());
        return nullptr;
    }

    // Create ImageView (grayscale)
    ImageView view(img_data, width, height, ImageFormat::Lum);

    // Configure reader options
    ReaderOptions reader_opts;
    if (options) {
        reader_opts.setTryHarder(options->try_harder != 0);
        reader_opts.setTryRotate(options->try_rotate != 0);
    }

    // Decode
    Barcodes barcodes = ReadBarcodes(view, reader_opts);

    stbi_image_free(img_data);

    if (barcodes.empty()) {
        set_error("No barcodes found");
        return nullptr;
    }

    // Return first result
    const Barcode& first = barcodes.front();
    DecodeResult* result = (DecodeResult*)malloc(sizeof(DecodeResult));
    if (!result) {
        set_error("Failed to allocate result");
        return nullptr;
    }

    std::string text = first.text();
    result->text = (char*)malloc(text.size() + 1);
    if (result->text) {
        strcpy(result->text, text.c_str());
    }
    result->format = convert_format(first.format());
    result->confidence = 1.0f;

    return result;
}

// Decodes tightly packed raw pixels without an intermediate encoded image.
EXPORT DecodeResult* decode_barcode_pixels(const unsigned char* data, int width, int height,
                                           int channels, const DecodeOptions* options) {
    if (!data || width <= 0 || height <= 0) {
        set_error("Invalid raw image data or dimensions");
        return nullptr;
    }

    auto formats = std::array{ImageFormat::None, ImageFormat::Lum, ImageFormat::LumA,
                              ImageFormat::RGB, ImageFormat::RGBA};
    if (channels < 1 || channels >= static_cast<int>(formats.size())) {
        set_error("Unsupported channel count: %d", channels);
        return nullptr;
    }

    try {
        ImageView view(data, width, height, formats.at(channels));
        ReaderOptions reader_opts;
        if (options) {
            reader_opts.setTryHarder(options->try_harder != 0);
            reader_opts.setTryRotate(options->try_rotate != 0);
            reader_opts.setTryInvert(options->try_invert != 0);
            reader_opts.setTryDownscale(options->try_downscale != 0);

            std::vector<ZXing::BarcodeFormat> selected_formats;
            if (options->formats != FORMAT_ALL && options->formats != FORMAT_NONE) {
                if (options->formats & FORMAT_QR_CODE) selected_formats.push_back(ZXing::BarcodeFormat::QRCode);
                if (options->formats & FORMAT_AZTEC) selected_formats.push_back(ZXing::BarcodeFormat::Aztec);
                if (options->formats & FORMAT_CODABAR) selected_formats.push_back(ZXing::BarcodeFormat::Codabar);
                if (options->formats & FORMAT_CODE_39) selected_formats.push_back(ZXing::BarcodeFormat::Code39);
                if (options->formats & FORMAT_CODE_93) selected_formats.push_back(ZXing::BarcodeFormat::Code93);
                if (options->formats & FORMAT_CODE_128) selected_formats.push_back(ZXing::BarcodeFormat::Code128);
                if (options->formats & FORMAT_DATA_MATRIX) selected_formats.push_back(ZXing::BarcodeFormat::DataMatrix);
                if (options->formats & FORMAT_EAN_8) selected_formats.push_back(ZXing::BarcodeFormat::EAN8);
                if (options->formats & FORMAT_EAN_13) selected_formats.push_back(ZXing::BarcodeFormat::EAN13);
                if (options->formats & FORMAT_ITF) selected_formats.push_back(ZXing::BarcodeFormat::ITF);
                if (options->formats & FORMAT_MAXICODE) selected_formats.push_back(ZXing::BarcodeFormat::MaxiCode);
                if (options->formats & FORMAT_PDF_417) selected_formats.push_back(ZXing::BarcodeFormat::PDF417);
                if (options->formats & FORMAT_UPC_A) selected_formats.push_back(ZXing::BarcodeFormat::UPCA);
                if (options->formats & FORMAT_UPC_E) selected_formats.push_back(ZXing::BarcodeFormat::UPCE);
                reader_opts.setFormats(ZXing::BarcodeFormats(std::move(selected_formats)));
            }
        }

        auto barcode = ReadBarcode(view, reader_opts);
        if (!barcode.isValid()) {
            set_error("No barcode found");
            return nullptr;
        }

        auto* result = new DecodeResult();
        result->text = strdup(barcode.text().c_str());
        if (!result->text) {
            delete result;
            set_error("Failed to allocate result text");
            return nullptr;
        }
        result->format = convert_format(barcode.format());
        result->confidence = 1.0f;
        return result;
    } catch (const std::exception& e) {
        set_error("Decode error: %s", e.what());
        return nullptr;
    }
}

// Empty main function required by Emscripten STANDALONE_WASM linker
// Only include for WASM builds to avoid symbol conflict with CGO
#ifdef __EMSCRIPTEN__
int main() {
    return 0;
}
#endif
