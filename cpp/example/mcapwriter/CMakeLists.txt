cmake_minimum_required(VERSION 3.20)
project(mcapwriter CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Protobuf REQUIRED)
add_subdirectory(../../include/filedesc ${CMAKE_CURRENT_BINARY_DIR}/filedesc)

# mcap uses lz4/zstd for chunk compression.
find_package(PkgConfig REQUIRED)
pkg_check_modules(LZ4 REQUIRED IMPORTED_TARGET liblz4)
pkg_check_modules(ZSTD REQUIRED IMPORTED_TARGET libzstd)

# Fetch the foxglove/mcap C++ header-only library.
include(FetchContent)
FetchContent_Populate(mcap
    GIT_REPOSITORY https://github.com/foxglove/mcap.git
    GIT_TAG releases/cpp/v2.2.0
)

# Generated protobuf C++ sources. Run 'make protoc' in cpp/example/ first.
set(PROTO_GEN_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../proto)
file(GLOB_RECURSE FOXGLOVE_PB_SOURCES CONFIGURE_DEPENDS ${PROTO_GEN_DIR}/*.pb.cc)

if(NOT FOXGLOVE_PB_SOURCES)
    message(FATAL_ERROR
        "No generated Foxglove protobuf sources found in ${PROTO_GEN_DIR}.\n"
        "Run 'make protoc' from cpp/example/ first."
    )
endif()

add_executable(mcapwriter main.cpp ${FOXGLOVE_PB_SOURCES})

target_include_directories(mcapwriter PRIVATE
    ${mcap_SOURCE_DIR}/cpp/mcap/include
    ${PROTO_GEN_DIR}
)

target_link_libraries(mcapwriter PRIVATE
    filedesc
    PkgConfig::LZ4
    PkgConfig::ZSTD
)

# Silence pedantic warnings from protobuf-generated code.
target_compile_options(mcapwriter PRIVATE
    -Wno-unused-parameter
    -Wno-shadow
    -Wno-pedantic
)

