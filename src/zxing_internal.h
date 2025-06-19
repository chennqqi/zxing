#ifndef ZXING_INTERNAL_H
#define ZXING_INTERNAL_H

#include <stdint.h>

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

// 释放解码结果
void zxing_free_result(DecodeResultInternal* result);

// 释放多个解码结果
void zxing_free_results(DecodeResultInternal** results, int count);

#endif // ZXING_INTERNAL_H 