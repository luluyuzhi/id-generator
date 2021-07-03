# CMake 最低版本号要求
cmake_minimum_required (VERSION 2.8)

# 项目信息
project (id-generator)

set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_STANDARD 11)

add_subdirectory(core/snow SNOW_ALGORITHM)

install (TARGETS snow DESTINATION lib)