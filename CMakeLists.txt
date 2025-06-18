cmake_minimum_required(VERSION 3.10)
project(zxing)

# 设置 C++ 标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 尝试多种方式查找 ZXing 库
set(ZXING_FOUND FALSE)

# 方法1: 使用 find_package
find_package(ZXing QUIET)
if(ZXing_FOUND)
    set(ZXING_FOUND TRUE)
    message(STATUS "ZXing found via find_package")
endif()

# 方法2: 使用 pkg-config
if(NOT ZXING_FOUND)
    find_package(PkgConfig QUIET)
    if(PkgConfig_FOUND)
        # 尝试不同的包名
        foreach(pkg_name "zxing-cpp" "zxing" "libzxing")
            pkg_check_modules(PC_ZXING QUIET ${pkg_name})
            if(PC_ZXING_FOUND)
                set(ZXING_FOUND TRUE)
                message(STATUS "ZXing found via pkg-config: ${pkg_name}")
                break()
            endif()
        endforeach()
    endif()
endif()

# 方法3: 手动查找头文件和库文件
if(NOT ZXING_FOUND)
    # 常见的安装路径
    set(ZXING_SEARCH_PATHS
        "/usr"
        "/usr/local"
        "/opt"
        "/opt/local"
    )
    
    # 查找头文件
    find_path(ZXING_INCLUDE_DIR
        NAMES ZXing.h
        PATHS ${ZXING_SEARCH_PATHS}
        PATH_SUFFIXES include include/zxing-cpp include/zxing
        NO_DEFAULT_PATH
    )
    
    # 查找库文件
    find_library(ZXING_LIBRARY
        NAMES zxing zxing-cpp libzxing
        PATHS ${ZXING_SEARCH_PATHS}
        PATH_SUFFIXES lib lib64
        NO_DEFAULT_PATH
    )
    
    if(ZXING_INCLUDE_DIR AND ZXING_LIBRARY)
        set(ZXING_FOUND TRUE)
        message(STATUS "ZXing found manually: ${ZXING_INCLUDE_DIR}, ${ZXING_LIBRARY}")
        
        # 创建导入目标
        add_library(ZXing::ZXing UNKNOWN IMPORTED)
        set_target_properties(ZXing::ZXing PROPERTIES
            IMPORTED_LOCATION "${ZXING_LIBRARY}"
            INTERFACE_INCLUDE_DIRECTORIES "${ZXING_INCLUDE_DIR}"
        )
    endif()
endif()

# 如果所有方法都失败，使用 FetchContent 自动下载和构建
if(NOT ZXING_FOUND)
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