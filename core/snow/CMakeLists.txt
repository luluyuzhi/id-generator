aux_source_directory(src DIR_LIB_SRCS)
include_directories(include)

add_library (snow SHARED ${DIR_LIB_SRCS})