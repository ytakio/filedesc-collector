add_executable(protodesc_test test_protodesc.cpp)
target_link_libraries(protodesc_test PRIVATE protodesc GTest::gtest_main absl::log_initialize absl::check)

add_test(NAME protodesc_test COMMAND protodesc_test)

