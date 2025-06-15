cmake_minimum_required(VERSION 3.10)
project(zxing)

# 设置 C++ 标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 查找 zxing-cpp 包
find_package(ZXing REQUIRED)

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