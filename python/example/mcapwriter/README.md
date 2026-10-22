# Python MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` to collect protobuf `FileDescriptorSet` schemas
- `foxglove-sdk` (the official Foxglove SDK) to write the file
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Go `go/example/mcapwriter` example.

The `foxglove-sdk` package provides native `foxglove.messages.SceneUpdate` and
`foxglove.channels.SceneUpdateChannel` classes, so the SceneUpdate channel does
not require manual protobuf serialization or a generated `FileDescriptorSet`.
The `/timestamp` channel, on the other hand, uses a generic
`foxglove.Channel` with a `FileDescriptorSet` produced by
`filedesc-collector`.

## Run

```bash
cd python/example/mcapwriter
make run
```

Or directly:

```bash
cd python/example/mcapwriter
python main.py
```

This creates `output.mcap` in the current directory.

## Install dependencies

### Using pip

```bash
cd python/example/mcapwriter
pip install -e ../../
pip install -r requirements.txt
```

### Using uv

```bash
cd python/example/mcapwriter
uv run main.py
```

## Clean generated files

```bash
make clean
```
