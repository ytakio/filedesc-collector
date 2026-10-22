# C++ MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` (header-only) to collect protobuf `FileDescriptorSet` schemas
- `foxglove-sdk` (the official Foxglove SDK for C++) to write the file
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Go `go/example/mcapwriter` example.

The Foxglove SDK C++ provides native `foxglove::messages::SceneUpdate` and
`foxglove::messages::SceneUpdateChannel` classes, so the SceneUpdate channel
does not require manual protobuf serialization or a generated
`FileDescriptorSet`. The `/timestamp` channel, on the other hand, uses a
generic `foxglove::RawChannel` with a `FileDescriptorSet` produced by
`filedesc-collector`.

## Build

```bash
cd cpp/example/mcapwriter
make build
```

This runs CMake and fetches the Foxglove SDK release archive for the current
platform automatically.

## Run

```bash
make run
```

Or directly:

```bash
cd build
./mcapwriter
```

This creates `output.mcap` in the build directory.

## Requirements

- CMake 3.20+
- C++17 compiler
- Protobuf development libraries
- Abseil (required by modern Protobuf)
- Network access to fetch the Foxglove SDK release archive

On macOS with Homebrew:

```bash
brew install protobuf
```

On Ubuntu/Debian:

```bash
sudo apt-get install cmake libprotobuf-dev protobuf-compiler libabsl-dev
```

## Clean

```bash
make clean
```

