
# CMake 最低版本号要求
cmake_minimum_required (VERSION 2.8)
SET(CMAKE_SYSTEM_NAME Linux) 
SET(CMAKE_C_COMPILER /usr/bin/gcc)

# 项目信息
project (id-generator)

set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_STANDARD 11)

add_subdirectory(core/snow SNOW_ALGORITHM)
add_subdirectory(core/UidGenerator UID_GENERATOR)
install (TARGETS snow DESTINATION lib)
install (TARGETS uidgenerator DESTINATION lib)