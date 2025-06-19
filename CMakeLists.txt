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
    add_library(zxingwrapper SHARED src/zxing.cpp)
    target_include_directories(zxingwrapper
        PUBLIC
            ${CMAKE_CURRENT_SOURCE_DIR}/include
            ${ZXING_INCLUDE_DIRS}
        PRIVATE
            ${CMAKE_CURRENT_SOURCE_DIR}/src
            ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/build/_deps/stb-src
    )
    target_link_libraries(zxingwrapper
        PRIVATE
            ${ZXING_LIBRARIES}
    )
else()
    message(STATUS "ZXing not found in system, using local implementation...")
    
    add_library(zxingwrapper SHARED src/zxing.cpp)
    target_include_directories(zxingwrapper
        PUBLIC
            ${CMAKE_CURRENT_SOURCE_DIR}/include
        PRIVATE
            ${CMAKE_CURRENT_SOURCE_DIR}/src
            ${CMAKE_CURRENT_SOURCE_DIR}/zxing-cpp/build/_deps/stb-src
    )
    target_link_libraries(zxingwrapper
        PRIVATE
            ZXing
    )
endif()

# 安装规则
install(TARGETS zxingwrapper
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
)

install(FILES include/zxing.h
    DESTINATION include
) 