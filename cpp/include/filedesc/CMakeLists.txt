find_package(Protobuf REQUIRED)
find_package(absl REQUIRED)

add_library(filedesc INTERFACE)
target_include_directories(filedesc INTERFACE
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/..>
    $<INSTALL_INTERFACE:include>
)
target_link_libraries(filedesc INTERFACE
    protobuf::libprotobuf
    absl::log_internal_check_op
    absl::log_internal_message
)

