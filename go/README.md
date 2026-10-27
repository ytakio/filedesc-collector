# Go filedesc

Go package for collecting `FileDescriptorProto` messages for a given protobuf
type, including all transitive file dependencies.

## Requirements

- Go 1.27+
- `google.golang.org/protobuf`

## Installation

```bash
go get github.com/ytakio/filedesc-collector/go/filedesc
```

## Usage

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

## Run tests

```bash
cd go
make
```

Or directly with `go test`:

```bash
cd go
go test ./filedesc
```

## Example

See [`example/mcapwriter`](example/mcapwriter) for an end-to-end MCAP writer
example. Run `make example` from `go/` to build and run it; it produces
`go/example/output.mcap`.

See the repository [README.md](../README.md) for full cross-language
documentation.

