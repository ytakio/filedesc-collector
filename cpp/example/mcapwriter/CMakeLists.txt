cmake_minimum_required(VERSION 3.20)
project(mcapwriter CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Protobuf REQUIRED)
find_package(absl REQUIRED)

# Determine the Foxglove SDK release archive for the current platform.
set(FOXGLOVE_SDK_VERSION "v0.30.2")
if(CMAKE_SYSTEM_NAME STREQUAL "Darwin")
  if(CMAKE_SYSTEM_PROCESSOR STREQUAL "arm64" OR CMAKE_SYSTEM_PROCESSOR STREQUAL "aarch64")
    set(FOXGLOVE_SDK_TRIPLE "aarch64-apple-darwin")
    set(FOXGLOVE_SDK_SHA256 "74454dcedac00fab8f1dd5308f88b101f87d15bf67e99e53def8d5dd11d05a81")
  elseif(CMAKE_SYSTEM_PROCESSOR STREQUAL "x86_64")
    set(FOXGLOVE_SDK_TRIPLE "x86_64-apple-darwin")
    set(FOXGLOVE_SDK_SHA256 "b3a521bdc2d2959dc1975b6f16b21472160544abb482a92a7597153c86e2a144")
  else()
    message(FATAL_ERROR "Unsupported macOS architecture: ${CMAKE_SYSTEM_PROCESSOR}")
  endif()
elseif(CMAKE_SYSTEM_NAME STREQUAL "Linux")
  if(CMAKE_SYSTEM_PROCESSOR STREQUAL "aarch64" OR CMAKE_SYSTEM_PROCESSOR STREQUAL "arm64")
    set(FOXGLOVE_SDK_TRIPLE "aarch64-unknown-linux-gnu")
    set(FOXGLOVE_SDK_SHA256 "9e388b0374ef2c3aa152bcfd1d7f4a8105283abd39cb5ea4ae74619bcfeb7fd7")
  elseif(CMAKE_SYSTEM_PROCESSOR STREQUAL "x86_64")
    set(FOXGLOVE_SDK_TRIPLE "x86_64-unknown-linux-gnu")
    set(FOXGLOVE_SDK_SHA256 "71bf8ded6907f2abb528b7844d9d441d82b93e0679c88724fc1144e6567be2d9")
  else()
    message(FATAL_ERROR "Unsupported Linux architecture: ${CMAKE_SYSTEM_PROCESSOR}")
  endif()
else()
  message(FATAL_ERROR "Unsupported platform: ${CMAKE_SYSTEM_NAME}")
endif()

set(FOXGLOVE_SDK_URL
  "https://github.com/foxglove/foxglove-sdk/releases/download/sdk/${FOXGLOVE_SDK_VERSION}/foxglove-${FOXGLOVE_SDK_VERSION}-cpp-${FOXGLOVE_SDK_TRIPLE}.zip"
)

# Fetch the Foxglove SDK release archive.
include(FetchContent)
FetchContent_Declare(
  foxglove_sdk
  DOWNLOAD_EXTRACT_TIMESTAMP TRUE
  URL ${FOXGLOVE_SDK_URL}
  URL_HASH SHA256=${FOXGLOVE_SDK_SHA256}
)
FetchContent_MakeAvailable(foxglove_sdk)

# Load the SDK's CMake package config.
find_package(foxglove-sdk CONFIG REQUIRED HINTS "${foxglove_sdk_SOURCE_DIR}")

# Compile the SDK's C++ wrapper sources against our toolchain.
foxglove_sdk_add_cpp_library(foxglove_cpp TYPE STATIC)

# filedesc-collector header lives two directories above this example.
set(FILEDESC_INCLUDE_DIR ../../include)

add_executable(mcapwriter main.cpp)

target_include_directories(mcapwriter PRIVATE
  ${FILEDESC_INCLUDE_DIR}
)

target_link_libraries(mcapwriter PRIVATE
  foxglove_cpp
  protobuf::libprotobuf
  absl::log_internal_check_op
  absl::log_internal_message
)

# On macOS the Foxglove SDK static library links IOKit symbols.
if(APPLE)
  target_link_libraries(mcapwriter PRIVATE
    "-framework IOKit"
    "-framework Foundation"
  )
endif()

