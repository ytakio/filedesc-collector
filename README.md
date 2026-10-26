# filedesc-collector

Libraries that collect `FileDescriptorProto` messages for a given protobuf type,
including all transitive file dependencies.

Repository: `github.com/ytakio/filedesc-collector`

## Spec

- Input: a protobuf message type (or a file descriptor).
- Output: a list of `FileDescriptorProto` messages.
  - The list contains the file that defines the input type.
  - It also contains every file that the target file transitively depends on.
  - Dependencies appear before the file that depends on them.

## Go

Module path: `github.com/ytakio/filedesc-collector/go`
Package path: `github.com/ytakio/filedesc-collector/go/filedesc`

### Install

```bash
go get github.com/ytakio/filedesc-collector/go/filedesc
```

### Usage

```go
import (
    "fmt"

    "github.com/ytakio/filedesc-collector/go/filedesc"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
    collector := filedesc.NewCollector()
    protos, err := collector.CollectForMessage((*timestamppb.Timestamp)(nil).ProtoReflect())
    if err != nil {
        // handle error
    }
    for _, p := range protos {
        fmt.Println(p.GetName())
    }
}
```

Run tests:

```bash
cd go
make
```

Run the example:

```bash
cd go
make example
```

## Python

Package name: `filedesc-collector`
Import name: `filedesc`

### Install

From PyPI (when published):

```bash
pip install filedesc-collector
```

Directly from this repository:

```bash
pip install "filedesc-collector @ git+https://github.com/ytakio/filedesc-collector.git#subdirectory=python"
```

### Usage

```python
from google.protobuf.timestamp_pb2 import Timestamp
from filedesc import FileDescriptorSetCollector

coll = FileDescriptorSetCollector()
protos = coll.collect_for_message(Timestamp.DESCRIPTOR)
for p in protos:
    print(p.name)
```

Run tests:

```bash
cd python
pip install -e '.[dev]'
pytest
```

## C++

Header-only library. Include path: `filedesc/filedesc.hpp`

### Install

Use CMake's `FetchContent` to consume the library directly from this repository:

```cmake
include(FetchContent)

FetchContent_Declare(
  filedesc
  GIT_REPOSITORY https://github.com/ytakio/filedesc-collector.git
  GIT_TAG        main
  SOURCE_SUBDIR  cpp
)

FetchContent_MakeAvailable(filedesc)

target_link_libraries(your_target PRIVATE filedesc)
```

### Usage

```cpp
#include <google/protobuf/timestamp.pb.h>
#include "filedesc/filedesc.hpp"

filedesc::FileDescriptorSetCollector collector;
auto protos = collector.CollectForMessage(google::protobuf::Timestamp::descriptor());
if (!protos.has_value()) {
    // handle error
}
for (const auto& p : protos.value()) {
    std::cout << p.name() << "\n";
}
```

Build and test:

```bash
cd cpp
mkdir build && cd build
cmake ..
make
ctest
```

## Examples

The repository includes end-to-end MCAP writer examples in three languages.
Each example writes a small MCAP file using `filedesc-collector` to supply
protobuf `FileDescriptorSet` schemas and a popular MCAP writer library to create
the file.

- [`python/example/mcapwriter`](python/example/mcapwriter) — Python example
  using `mcap-protobuf-support` and `foxglove-schemas-protobuf`. Run `make run`
  from `python/` to produce `python/build/output.mcap`.
- [`cpp/example/mcapwriter`](cpp/example/mcapwriter) — C++ example using the
  Foxglove MCAP C++ library and generated Foxglove protobuf bindings. Run
  `make run` from `cpp/` to produce `cpp/build/example/mcapwriter/output.mcap`.
- [`go/example/mcapwriter`](go/example/mcapwriter) — Go example using the
  Foxglove MCAP Go SDK and generated Foxglove protobuf bindings. Run `make run`
  from `go/` to produce `go/build/output.mcap`.

## Notes

- Dependencies are collected in depth-first order and deduplicated within a
  single collection. A `FileDescriptorSetCollector` instance can be reused for
  multiple independent collections.
- If you need a single combined `FileDescriptorSet`, wrap the returned protos in
  one:

  ```go
  fds := &descriptorpb.FileDescriptorSet{File: protos}
  ```

