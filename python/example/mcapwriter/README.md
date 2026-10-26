# Python MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` to collect protobuf `FileDescriptorSet` schemas
- the `mcap` Python library to write the file
- generated `foxglove` protobuf bindings under `python/proto/foxglove/` from the
  upstream `.proto` schemas
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Go and C++ `mcapwriter` examples.

## Run

From the `python/` directory:

```bash
make example
```

Or directly with `uv`:

```bash
cd python
uv run python example/mcapwriter/main.py
```

This downloads the Foxglove SDK protobuf schemas, extracts them into `proto/`
at the repository root, generates Python protobuf bindings into
`python/proto/`, and writes `python/example/mcapwriter/output.mcap`.

Ensure the package is installed in the active virtual environment (e.g.
`uv pip install -e .`) so `filedesc` and `proto` can be imported when running a
script by path.

## Clean generated files

```bash
make clean
```

To also remove the shared schema archive and extracted `.proto` files from the
repository root, run:

```bash
make clean-all
```
