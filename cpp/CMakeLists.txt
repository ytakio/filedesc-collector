cmake_minimum_required(VERSION 3.16)
project(filedesc CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Protobuf REQUIRED)
find_package(absl REQUIRED)

# Header-only library.
add_library(filedesc INTERFACE)

target_include_directories(filedesc
    INTERFACE
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(filedesc INTERFACE protobuf::libprotobuf)

option(BUILD_TESTING "Build tests" ON)
if(BUILD_TESTING)
    find_package(GTest)
    if(GTest_FOUND)
        enable_testing()
        add_subdirectory(tests)
    else()
        message(STATUS "GTest not found; skipping tests")
    endif()
endif()

