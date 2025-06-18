cmake_minimum_required(VERSION 3.10)
project(zxing)

# 设置 C++ 标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 尝试查找系统安装的 ZXing 库
find_package(ZXing QUIET)

# 如果系统没有安装 ZXing，则使用 FetchContent 自动下载和构建
if(NOT ZXing_FOUND)
    message(STATUS "ZXing not found in system, downloading and building from source...")
    include(FetchContent)
    FetchContent_Declare(
        zxing-cpp
        GIT_REPOSITORY https://github.com/zxing-cpp/zxing-cpp.git
        GIT_TAG v2.2.0
        GIT_SHALLOW TRUE
    )
    FetchContent_MakeAvailable(zxing-cpp)
else()
    message(STATUS "Using system-installed ZXing library")
endif()

# 添加库目标
add_library(zxing SHARED
    src/zxing.cpp
)

# 设置包含目录
target_include_directories(zxing
    PUBLIC
        ${CMAKE_CURRENT_SOURCE_DIR}/include
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# 链接依赖库
target_link_libraries(zxing
    PRIVATE
        ZXing::ZXing
)

# 安装规则
install(TARGETS zxing
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(FILES include/zxing.h
    DESTINATION include
) 