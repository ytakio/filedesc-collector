add_executable(filedesc_test test_filedesc.cpp)
target_link_libraries(filedesc_test PRIVATE filedesc GTest::gtest_main absl::log_initialize absl::check)

add_test(NAME filedesc_test COMMAND filedesc_test)

