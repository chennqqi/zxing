cmake_minimum_required(VERSION 3.10)
project(zxing)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 选项：是否编译静态库（默认开启，符合项目需求）
option(BUILD_STATIC_LIB "Build static library instead of shared library" ON)

# 设置输出目录
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# 添加ZXingCPP子目录，确保ZXingCPP也编译为静态库
if(BUILD_STATIC_LIB)
    set(BUILD_SHARED_LIB_ZXING OFF CACHE BOOL "Build ZXingCPP as static library" FORCE)
else()
    set(BUILD_SHARED_LIB_ZXING ON CACHE BOOL "Build ZXingCPP as shared library" FORCE)
endif()

# 添加ZXingCPP子目录
add_subdirectory(zxing-cpp)

# 创建wrapper库（根据选项选择静态库或动态库）
if(BUILD_STATIC_LIB)
    add_library(zxingwrapper STATIC src/zxing.cpp)
else()
    add_library(zxingwrapper SHARED src/zxing.cpp)
endif()

# 设置包含目录
target_include_directories(zxingwrapper
    PUBLIC
        ${CMAKE_CURRENT_SOURCE_DIR}/include
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src/oned
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src/qrcode
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src/datamatrix
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src/aztec
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src/pdf417
        ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/core/src/maxicode
)

# 链接ZXing库
target_link_libraries(zxingwrapper
    PRIVATE
        ZXing
)

# 安装规则
install(TARGETS zxingwrapper
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
)

install(FILES include/zxing.h
    DESTINATION include
) 