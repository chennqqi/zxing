#ifndef ZXING_H
#define ZXING_H

#ifdef __cplusplus
extern "C" {
#endif

// 条码格式枚举
typedef enum {
    FORMAT_NONE = 0,
    FORMAT_QR_CODE = 1,
    FORMAT_AZTEC = 2,
    FORMAT_CODABAR = 4,
    FORMAT_CODE_39 = 8,
    FORMAT_CODE_93 = 16,
    FORMAT_CODE_128 = 32,
    FORMAT_DATA_MATRIX = 64,
    FORMAT_EAN_8 = 128,
    FORMAT_EAN_13 = 256,
    FORMAT_ITF = 512,
    FORMAT_MAXICODE = 1024,
    FORMAT_PDF_417 = 2048,
    FORMAT_UPC_A = 4096,
    FORMAT_UPC_E = 8192,
    FORMAT_ALL = 0xFFFF
} BarcodeFormat;

// 解码结果结构体
typedef struct {
    char* text;           // 解码文本
    BarcodeFormat format; // 条码格式
    float confidence;     // 置信度
} DecodeResult;

// 解码选项结构体
typedef struct {
    BarcodeFormat formats;    // 要识别的条码格式
    int try_harder;          // 是否尝试更努力的解码
    int try_rotate;          // 是否尝试旋转图像
    int try_invert;          // 是否尝试反转图像
    int try_downscale;       // 是否尝试缩小图像
} DecodeOptions;

// 初始化默认解码选项
DecodeOptions* create_default_options();

// 释放解码选项
void free_options(DecodeOptions* options);

// 识别单个条码
DecodeResult* decode_barcode(const char* image_path, const DecodeOptions* options);

// 识别多个条码
DecodeResult** decode_barcodes(const char* image_path, const DecodeOptions* options, int* count);

// 释放单个解码结果
void free_result(DecodeResult* result);

// 释放多个解码结果
void free_results(DecodeResult** results, int count);

// 获取错误信息
const char* get_last_error();

#ifdef __cplusplus
}
#endif

#endif // ZXING_H 