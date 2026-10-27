# C++ MCAP writer example

This example writes a small MCAP file using:

- `filedesc-collector` (header-only) to collect protobuf
  `FileDescriptorSet` schemas
- the [Foxglove MCAP C++ library](https://github.com/foxglove/mcap) to write
  the file
- generated `foxglove` protobuf bindings under `../proto/foxglove/` from the
  upstream `.proto` schemas

This mirrors the Go and Python `mcapwriter` examples.

## Build and run

From this directory (`cpp/example/mcapwriter/`):

```bash
cd ../..
make example
```

Or build and run directly:

```bash
cd ..
cmake -B build -S mcapwriter
cmake --build build
./build/mcapwriter
```

This downloads the Foxglove SDK protobuf schemas, extracts them into `proto/`
at the repository root, generates C++ protobuf bindings into `../proto/`,
builds the example under `../build/`, and writes `../output.mcap`.

## Requirements

- CMake 3.20+
- C++17 compiler
- Protobuf development libraries (tested with libprotoc 7.36.2 / protobuf 36.2)
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

From this directory:

```bash
cd ../..
make clean
```

This removes the generated `../build/` and `../proto/` directories.

To also remove the shared schema archive and extracted `.proto` files from the
repository root, run:

```bash
cd ../..
make clean-all
```
