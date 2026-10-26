cmake_minimum_required(VERSION 3.20)
project(mcapwriter CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Protobuf REQUIRED)
find_package(absl REQUIRED)

# The foxglove/mcap C++ library is header-only but depends on lz4 and zstd for
# chunk compression. Try pkg-config first, then fall back to well-known paths.
find_package(PkgConfig)
if(PkgConfig_FOUND)
  pkg_check_modules(LZ4 IMPORTED_TARGET liblz4)
  pkg_check_modules(ZSTD IMPORTED_TARGET libzstd)
endif()

if(NOT TARGET PkgConfig::LZ4)
  find_library(LZ4_LIBRARY NAMES lz4 PATHS /opt/homebrew/lib /usr/local/lib)
  find_path(LZ4_INCLUDE_DIR NAMES lz4frame.h PATHS /opt/homebrew/include /usr/local/include)
  if(NOT LZ4_LIBRARY OR NOT LZ4_INCLUDE_DIR)
    message(FATAL_ERROR "lz4 not found. Install it (e.g. brew install lz4 or apt install liblz4-dev)")
  endif()
  add_library(PkgConfig::LZ4 UNKNOWN IMPORTED)
  set_target_properties(PkgConfig::LZ4 PROPERTIES
    IMPORTED_LOCATION ${LZ4_LIBRARY}
    INTERFACE_INCLUDE_DIRECTORIES ${LZ4_INCLUDE_DIR}
  )
endif()

if(NOT TARGET PkgConfig::ZSTD)
  find_library(ZSTD_LIBRARY NAMES zstd PATHS /opt/homebrew/lib /usr/local/lib)
  find_path(ZSTD_INCLUDE_DIR NAMES zstd.h PATHS /opt/homebrew/include /usr/local/include)
  if(NOT ZSTD_LIBRARY OR NOT ZSTD_INCLUDE_DIR)
    message(FATAL_ERROR "zstd not found. Install it (e.g. brew install zstd or apt install libzstd-dev)")
  endif()
  add_library(PkgConfig::ZSTD UNKNOWN IMPORTED)
  set_target_properties(PkgConfig::ZSTD PROPERTIES
    IMPORTED_LOCATION ${ZSTD_LIBRARY}
    INTERFACE_INCLUDE_DIRECTORIES ${ZSTD_INCLUDE_DIR}
  )
endif()

# Fetch the foxglove/mcap C++ header-only library. It does not have a
# CMakeLists.txt at its root, so we only populate the source tree and use the
# include directory directly.
include(FetchContent)
if(POLICY CMP0169)
  # FetchContent_Populate with declared details is deprecated in CMake 3.30+,
  # but mcap has no top-level CMakeLists.txt for FetchContent_MakeAvailable.
  cmake_policy(SET CMP0169 OLD)
endif()
FetchContent_Declare(
  mcap
  GIT_REPOSITORY https://github.com/foxglove/mcap.git
  GIT_TAG releases/cpp/v2.2.0
)
FetchContent_GetProperties(mcap)
if(NOT mcap_POPULATED)
  FetchContent_Populate(mcap)
endif()
set(MCAP_INCLUDE_DIR ${mcap_SOURCE_DIR}/cpp/mcap/include)

# filedesc-collector header lives two directories above this example.
set(FILEDESC_INCLUDE_DIR ../../include)

# Generated protobuf C++ sources. The Makefile downloads the .proto files and
# runs protoc before invoking CMake; the layout under $(PROTO_GEN_DIR) mirrors
# the layout of the upstream schema directory, so glob recursively rather than
# hardcoding a subdirectory such as foxglove/.
set(PROTO_GEN_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../../proto)
file(GLOB_RECURSE FOXGLOVE_PB_SOURCES CONFIGURE_DEPENDS ${PROTO_GEN_DIR}/*.pb.cc)
file(GLOB_RECURSE FOXGLOVE_PB_HEADERS CONFIGURE_DEPENDS ${PROTO_GEN_DIR}/*.pb.h)

add_executable(mcapwriter main.cpp ${FOXGLOVE_PB_SOURCES})

target_include_directories(mcapwriter PRIVATE
  ${FILEDESC_INCLUDE_DIR}
  ${MCAP_INCLUDE_DIR}
  ${PROTO_GEN_DIR}
)

target_link_libraries(mcapwriter PRIVATE
  protobuf::libprotobuf
  PkgConfig::LZ4
  PkgConfig::ZSTD
  absl::log_internal_check_op
  absl::log_internal_message
)

# Protobuf-generated code triggers a number of pedantic warnings; silence them
# for this example so -Werror in a parent project does not break the build.
if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang|AppleClang")
  target_compile_options(mcapwriter PRIVATE
    -Wno-unused-parameter
    -Wno-shadow
    -Wno-pedantic
  )
endif()

