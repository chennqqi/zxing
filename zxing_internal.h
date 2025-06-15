#ifndef ZXING_INTERNAL_H
#define ZXING_INTERNAL_H

#include <stdint.h>

// 图像格式
typedef enum {
    IMAGE_FORMAT_UNKNOWN = 0,
    IMAGE_FORMAT_LUMINANCE = 1,
    IMAGE_FORMAT_RGB = 2,
    IMAGE_FORMAT_RGBA = 3,
    IMAGE_FORMAT_BGR = 4,
    IMAGE_FORMAT_BGRA = 5
} ImageFormat;

// 图像结构体
typedef struct {
    uint8_t* data;        // 图像数据
    int width;            // 宽度
    int height;           // 高度
    ImageFormat format;   // 格式
    int stride;           // 行跨度
} Image;

// 解码结果结构体
typedef struct {
    char* text;           // 解码文本
    int format;           // 条码格式
    float confidence;     // 置信度
    int x1, y1;           // 左上角坐标
    int x2, y2;           // 右上角坐标
    int x3, y3;           // 右下角坐标
    int x4, y4;           // 左下角坐标
} DecodeResultInternal;

// 初始化 zxing-cpp
int zxing_init();

// 清理 zxing-cpp
void zxing_cleanup();

// 加载图像
Image* zxing_load_image(const char* path);

// 释放图像
void zxing_free_image(Image* image);

// 解码单个条码
DecodeResultInternal* zxing_decode(const Image* image, int formats, int try_harder, int try_rotate, int try_invert, int try_downscale);

// 解码多个条码
DecodeResultInternal** zxing_decode_multi(const Image* image, int formats, int try_harder, int try_rotate, int try_invert, int try_downscale, int* count);

// 释放解码结果
void zxing_free_result(DecodeResultInternal* result);

// 释放多个解码结果
void zxing_free_results(DecodeResultInternal** results, int count);

#endif // ZXING_INTERNAL_H 