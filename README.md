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

### Go installation

```bash
go get github.com/ytakio/filedesc-collector/go/filedesc
```

### Go usage

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

### Python installation

From PyPI (when published):

```bash
pip install filedesc-collector
```

Directly from this repository:

```bash
pip install "filedesc-collector @ git+https://github.com/ytakio/filedesc-collector.git#subdirectory=python"
```

The project uses `uv` for local development. To install in editable mode with
`uv`:

```bash
cd python
uv pip install -e '.[dev]'
```

The `proto/` package used by the example is generated at build time; it is not
required for the core library or tests.

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
make
```

Or directly with `uv`:

```bash
cd python
uv run pytest
```

## C++

Header-only library. Include path: `filedesc/filedesc.hpp`

### C++ installation

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

### C++ usage

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

### Build and test

```bash
cd cpp
make
```

To build only the library and tests without examples:

```bash
cd cpp
cmake -B build -S . -DBUILD_EXAMPLES=OFF
make
ctest
```

## Examples

The repository includes end-to-end MCAP writer examples in three languages.
Each example writes a small MCAP file using `filedesc-collector` to supply
protobuf `FileDescriptorSet` schemas and a popular MCAP writer library to create
the file.

- [`python/example/mcapwriter.py`](python/example/mcapwriter.py) — Python
  example using the `mcap` Python library and generated Foxglove protobuf
  bindings. Run `make example` from `python/` to produce
  `python/example/output.mcap`.
- [`cpp/example/mcapwriter/main.cpp`](cpp/example/mcapwriter/main.cpp) — C++
  example using the Foxglove MCAP C++ library and generated Foxglove protobuf
  bindings. Run `make example` from `cpp/` to produce
  `cpp/example/output.mcap`.
- [`go/example/mcapwriter/main.go`](go/example/mcapwriter/main.go) — Go
  example using the Foxglove MCAP Go SDK and generated Foxglove protobuf
  bindings. Run `make example` from `go/` to produce
  `go/example/output.mcap`.

## Repository layout

```text
.
├── go/              # Go library and example
├── python/          # Python library and example
├── cpp/             # C++ header-only library and example
├── .github/         # GitHub Actions CI workflows
├── Makefile         # Top-level orchestration for all languages
├── LICENSE          # MIT License
└── README.md
```

## Notes

- Dependencies are collected in depth-first order and deduplicated within a
  single collection. A `FileDescriptorSetCollector` instance can be reused for
  multiple independent collections.
- If you need a single combined `FileDescriptorSet`, wrap the returned protos
  in one: `fds := &descriptorpb.FileDescriptorSet{File: protos}`.
