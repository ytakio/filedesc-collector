find_package(Protobuf REQUIRED)

add_library(filedesc INTERFACE)
target_include_directories(filedesc INTERFACE
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/..>
    $<INSTALL_INTERFACE:include>
)
target_link_libraries(filedesc INTERFACE
    protobuf::libprotobuf
)

# Newer Protobuf builds (e.g. Homebrew's) depend on absl logging internals but
# do not always re-export them transitively. Try to find absl and link the
# targets when they are available.
find_package(absl QUIET)
if(TARGET absl::log_internal_check_op)
    target_link_libraries(filedesc INTERFACE absl::log_internal_check_op)
endif()
if(TARGET absl::log_internal_message)
    target_link_libraries(filedesc INTERFACE absl::log_internal_message)
endif()

