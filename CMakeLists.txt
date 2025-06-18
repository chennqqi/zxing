cmake_minimum_required(VERSION 3.10)
project(zxing)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(PkgConfig QUIET)
if(PkgConfig_FOUND)
    pkg_check_modules(ZXING zxing)
endif()

if(ZXING_FOUND)
    message(STATUS "ZXing found via pkg-config")
    add_library(zxing SHARED src/zxing.cpp)
    target_include_directories(zxing
        PUBLIC
            ${CMAKE_CURRENT_SOURCE_DIR}/include
            ${ZXING_INCLUDE_DIRS}
        PRIVATE
            ${CMAKE_CURRENT_SOURCE_DIR}/src
    )
    target_link_libraries(zxing
        PRIVATE
            ${ZXING_LIBRARIES}
    )
else()
    message(STATUS "ZXing not found in system, using local zxing-cpp subdirectory...")
    add_subdirectory(zxing-cpp)
    add_library(zxing SHARED src/zxing.cpp)
    target_include_directories(zxing
        PUBLIC
            ${CMAKE_CURRENT_SOURCE_DIR}/include
        PRIVATE
            ${CMAKE_CURRENT_SOURCE_DIR}/src
    )
    target_link_libraries(zxing
        PRIVATE
            ZXing::ZXing
    )
endif()

# 安装规则
install(TARGETS zxing
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
)

install(FILES include/zxing.h
    DESTINATION include
) 