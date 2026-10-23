# C++ MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` (header-only) to collect protobuf
  `FileDescriptorSet` schemas
- the [Foxglove MCAP C++ library](https://github.com/foxglove/mcap) to write
  the file
- `google.protobuf.Timestamp` and `foxglove.SceneUpdate` as sample messages

This mirrors the Go and Python `mcapwriter` examples.

## Build

```bash
cd cpp/example/mcapwriter
make build
```

This downloads the Foxglove SDK protobuf schemas, extracts them into
`proto/foxglove/` at the repository root, generates C++ protobuf bindings into
`cpp/proto/foxglove/`, and then builds the example with CMake.

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
- Protobuf development libraries (tested with 7.36.2 / protobuf 36.2)
- Abseil (required by modern Protobuf)
- lz4 and zstd libraries (required by the MCAP writer for chunk compression)
- Network access to fetch the foxglove/mcap C++ headers and Foxglove schema tarball

On macOS with Homebrew:

```bash
brew install protobuf lz4 zstd pkg-config
```

On Ubuntu/Debian:

```bash
sudo apt-get install cmake libprotobuf-dev protobuf-compiler libabsl-dev \
  liblz4-dev libzstd-dev pkg-config
```

## Clean

```bash
make clean
```

This removes the build directory and the generated `proto/` directory under
this example.

To also remove the shared schema archive and extracted `.proto` files from the
repository root, run:

```bash
make clean-all
```
