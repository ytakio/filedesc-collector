# Python MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` to collect protobuf `FileDescriptorSet` schemas
- the `mcap` Python library to write the file
- generated `foxglove` protobuf bindings from the upstream `.proto` schemas
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Go `go/example/mcapwriter` example.

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

This creates `output.mcap` in `python/`. Ensure the package is installed in the
active virtual environment (e.g. `uv pip install -e .`) so `filedesc` can be
imported when running a script by path.

## Clean generated files

```bash
make clean
```
