# Python MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` to collect protobuf `FileDescriptorSet` schemas
- the `mcap` Python library to write the file
- generated `foxglove` protobuf bindings under `./proto/foxglove/` from the
  upstream `.proto` schemas
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Go and C++ `mcapwriter` examples.

## Run

From this directory (`python/example/`):

```bash
make example
```

Or run the script directly:

```bash
uv run python mcapwriter.py
```

This downloads the Foxglove SDK protobuf schemas, extracts them into `proto/`
at the repository root, generates Python protobuf bindings into `./proto/`,
and writes `./output.mcap`.

Ensure the package is installed in the active virtual environment (e.g.
`uv pip install -e .`) so `filedesc` and `proto` can be imported when running a
script by path.

## Clean generated files

From this directory:

```bash
make clean
```

To also remove the shared schema archive and extracted `.proto` files from the
repository root, run:

```bash
make clean-all
```
