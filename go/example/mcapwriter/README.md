# Go MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` to collect protobuf `FileDescriptorSet` schemas
- the Foxglove MCAP Go SDK (`github.com/foxglove/mcap/go/mcap`) to write the
  file
- generated `foxglove` protobuf bindings under `../proto/foxglove/` from the
  upstream `.proto` schemas
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Python and C++ `mcapwriter` examples.

## Run

From this directory (`go/example/mcapwriter/`):

```bash
cd ../..
make example
```

Or run the example directly:

```bash
cd ..
go run ./mcapwriter
```

This downloads the Foxglove SDK protobuf schemas, extracts them into `proto/`
at the repository root, generates Go protobuf bindings into `../proto/`, and
writes `../output.mcap`.

## Clean generated files

From this directory:

```bash
cd ../..
make clean
```

To also remove the shared schema archive and extracted `.proto` files from the
repository root, run:

```bash
cd ../..
make clean-all
```
