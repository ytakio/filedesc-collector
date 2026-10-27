find_package(GTest REQUIRED)

add_executable(filedesc_test test_filedesc.cpp)
target_link_libraries(filedesc_test PRIVATE
    filedesc
    GTest::gtest_main
)

include(GoogleTest)
gtest_discover_tests(filedesc_test)

