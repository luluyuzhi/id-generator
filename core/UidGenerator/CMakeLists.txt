aux_source_directory(src DIR_LIB_SRCS)
include_directories(include)
set(CMAKE_C_STANDARD 11)
add_library (uidgenerator SHARED ${DIR_LIB_SRCS})