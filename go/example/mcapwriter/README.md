# Go MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` to collect protobuf `FileDescriptorSet` schemas
- the Foxglove MCAP Go SDK (`github.com/foxglove/mcap/go/mcap`) to write the
  file
- generated `foxglove` protobuf bindings under `go/proto/foxglove/` from the
  upstream `.proto` schemas
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Python and C++ `mcapwriter` examples.

## Run

From the `go/` directory:

```bash
make example
```

Or directly:

```bash
cd go
go run ./example/mcapwriter
```

This downloads the Foxglove SDK protobuf schemas, extracts them into `proto/`
at the repository root, generates Go protobuf bindings into `go/proto/`, and
writes `go/output.mcap`.

## Clean generated files

```bash
make clean
```

To also remove the shared schema archive and extracted `.proto` files from the
repository root, run:

```bash
make clean-all
```
