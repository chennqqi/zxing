#include "zxing.h"
#include "zxing_internal.h"
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>

// 错误信息缓冲区
static char last_error[256] = {0};

// 设置错误信息
static void set_error(const char* format, ...) {
    va_list args;
    va_start(args, format);
    vsnprintf(last_error, sizeof(last_error), format, args);
    va_end(args);
}

// 初始化 zxing
static int initialized = 0;

// 创建默认解码选项
DecodeOptions* create_default_options() {
    DecodeOptions* options = (DecodeOptions*)malloc(sizeof(DecodeOptions));
    if (!options) {
        set_error("Failed to allocate memory for options");
        return NULL;
    }
    
    // 设置默认值
    options->formats = FORMAT_ALL;  // 支持所有格式
    options->try_harder = 1;       // 默认尝试更努力的解码
    options->try_rotate = 1;       // 默认尝试旋转
    options->try_invert = 0;       // 默认不尝试反转
    options->try_downscale = 1;    // 默认尝试缩小
    
    return options;
}

// 释放解码选项
void free_options(DecodeOptions* options) {
    if (options) {
        free(options);
    }
}

// 释放单个解码结果
void free_result(DecodeResult* result) {
    if (result) {
        if (result->text) {
            free(result->text);
        }
        free(result);
    }
}

// 释放多个解码结果
void free_results(DecodeResult** results, int count) {
    if (results) {
        for (int i = 0; i < count; i++) {
            free_result(results[i]);
        }
        free(results);
    }
}

// 获取错误信息
const char* get_last_error() {
    return last_error;
}

// 转换内部解码结果到外部结果
static DecodeResult* convert_result(DecodeResultInternal* internal) {
    if (!internal) return NULL;
    
    DecodeResult* result = (DecodeResult*)malloc(sizeof(DecodeResult));
    if (!result) {
        set_error("Failed to allocate memory for result");
        return NULL;
    }
    
    result->text = strdup(internal->text);
    result->format = internal->format;
    result->confidence = internal->confidence;
    
    return result;
}

// 识别单个条码
DecodeResult* decode_barcode(const char* image_path, const DecodeOptions* options) {
    if (!image_path) {
        set_error("Image path is NULL");
        return NULL;
    }
    
    if (!options) {
        set_error("Options is NULL");
        return NULL;
    }
    
    // 确保已初始化
    if (!initialized) {
        if (zxing_init() != 0) {
            set_error("Failed to initialize zxing");
            return NULL;
        }
        initialized = 1;
    }
    
    // 加载图像
    Image* image = zxing_load_image(image_path);
    if (!image) {
        set_error("Failed to load image: %s", image_path);
        return NULL;
    }
    
    // 解码
    DecodeResultInternal* internal = zxing_decode(
        image,
        options->formats,
        options->try_harder,
        options->try_rotate,
        options->try_invert,
        options->try_downscale
    );
    
    // 释放图像
    zxing_free_image(image);
    
    if (!internal) {
        set_error("Failed to decode barcode");
        return NULL;
    }
    
    // 转换结果
    DecodeResult* result = convert_result(internal);
    
    // 释放内部结果
    zxing_free_result(internal);
    
    return result;
}

// 识别多个条码
DecodeResult** decode_barcodes(const char* image_path, const DecodeOptions* options, int* count) {
    if (!image_path) {
        set_error("Image path is NULL");
        return NULL;
    }
    
    if (!options) {
        set_error("Options is NULL");
        return NULL;
    }
    
    if (!count) {
        set_error("Count pointer is NULL");
        return NULL;
    }
    
    // 确保已初始化
    if (!initialized) {
        if (zxing_init() != 0) {
            set_error("Failed to initialize zxing");
            return NULL;
        }
        initialized = 1;
    }
    
    // 加载图像
    Image* image = zxing_load_image(image_path);
    if (!image) {
        set_error("Failed to load image: %s", image_path);
        return NULL;
    }
    
    // 解码多个条码
    DecodeResultInternal** internals = zxing_decode_multi(
        image,
        options->formats,
        options->try_harder,
        options->try_rotate,
        options->try_invert,
        options->try_downscale,
        count
    );
    
    // 释放图像
    zxing_free_image(image);
    
    if (!internals) {
        set_error("Failed to decode barcodes");
        return NULL;
    }
    
    // 分配结果数组
    DecodeResult** results = (DecodeResult**)malloc(sizeof(DecodeResult*) * (*count));
    if (!results) {
        set_error("Failed to allocate memory for results");
        zxing_free_results(internals, *count);
        return NULL;
    }
    
    // 转换每个结果
    for (int i = 0; i < *count; i++) {
        results[i] = convert_result(internals[i]);
        if (!results[i]) {
            // 转换失败，清理已分配的内存
            for (int j = 0; j < i; j++) {
                free_result(results[j]);
            }
            free(results);
            zxing_free_results(internals, *count);
            return NULL;
        }
    }
    
    // 释放内部结果
    zxing_free_results(internals, *count);
    
    return results;
} 