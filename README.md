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

Module path: `github.com/ytakio/filedesc-collector`
Package path: `github.com/ytakio/filedesc-collector/go/filedesc`

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
cd go/filedesc
go test ./...
```

## Python

Package name: `filedesc-collector`
Import name: `filedesc`

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

## Notes

- Dependencies are collected in depth-first order and deduplicated within a
  single collection. A `FileDescriptorSetCollector` instance can be reused for
  multiple independent collections.
- If you need a single combined `FileDescriptorSet`, wrap the returned protos in
  one:

  ```go
  fds := &descriptorpb.FileDescriptorSet{File: protos}
  ```

