# protodesc

Libraries that collect `FileDescriptorProto` messages for a given protobuf type,
including all transitive file dependencies.

Repository: `github.com/ytakio/protodesc`

## Spec

- Input: a protobuf message type (or a file descriptor).
- Output: a list of `FileDescriptorProto` messages.
  - The list contains the file that defines the input type.
  - It also contains every file that the target file transitively depends on.
  - Dependencies appear before the file that depends on them.

## Go

Module path: `github.com/ytakio/protodesc`
Package path: `github.com/ytakio/protodesc/go/protodesc`

```go
import (
    "fmt"

    "github.com/ytakio/protodesc/go/protodesc"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
    protos, err := protodesc.CollectForMessage((*timestamppb.Timestamp)(nil).ProtoReflect())
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
cd go/protodesc
go test ./...
```

## Python

Package name: `protodesc`

```python
from google.protobuf.timestamp_pb2 import Timestamp
from protodesc import collect_for_message

protos = collect_for_message(Timestamp.DESCRIPTOR)
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

Header-only library. Include path: `protodesc/protodesc.hpp`

```cpp
#include <google/protobuf/timestamp.pb.h>
#include "protodesc/protodesc.hpp"

auto protos = protodesc::CollectForMessage(google::protobuf::Timestamp::descriptor());
for (const auto& p : protos) {
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

- Dependencies are collected in depth-first order and deduplicated across a
  single collector instance.
- If you need a single combined `FileDescriptorSet`, wrap the returned protos in
  one:

  ```go
  fds := &descriptorpb.FileDescriptorSet{File: protos}
  ```

