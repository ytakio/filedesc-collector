# C++ filedesc

Header-only library.

## Requirements

- CMake 3.16+
- C++17 compiler
- Protobuf development libraries (tested with 7.36.2 / protobuf 36.2)
- Abseil (required by modern Protobuf)
- GTest (optional, for tests)

On macOS with Homebrew:

```bash
brew install protobuf googletest
```

## Build

Because this is a header-only library, there is no library binary to build. The
CMake target `filedesc` is an `INTERFACE` library that only adds the include
directory and links `protobuf::libprotobuf`.

```bash
cd cpp
mkdir -p build && cd build
cmake ..
make -j4
```

To use it in your own CMake project:

```cmake
add_subdirectory(filedesc)
target_link_libraries(your_target PRIVATE filedesc)
```

To disable tests:

```bash
cmake -DBUILD_TESTING=OFF ..
```

## Run tests

```bash
make
ctest --output-on-failure
```

## Usage

```cpp
#include <google/protobuf/timestamp.pb.h>
#include "filedesc/filedesc.hpp"

auto protos = filedesc::CollectForMessage(google::protobuf::Timestamp::descriptor());
for (const auto& p : protos) {
    std::cout << p.name() << "\n";
}
```

