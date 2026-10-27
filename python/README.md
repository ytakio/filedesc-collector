# filedesc

Python package for collecting `FileDescriptorProto` messages for a given
protobuf type, including all transitive file dependencies.

## Requirements

- Python 3.14+
- `protobuf` 7.36.2+

## Installation

From PyPI (when published):

```bash
pip install filedesc-collector
```

Directly from this repository:

```bash
pip install "filedesc-collector @ git+https://github.com/ytakio/filedesc-collector.git#subdirectory=python"
```

The project uses `uv` for local development. To install in editable mode with
dev dependencies:

```bash
cd python
uv pip install -e '.[dev]'
```

## Usage

```python
from google.protobuf.timestamp_pb2 import Timestamp
from filedesc import FileDescriptorSetCollector

coll = FileDescriptorSetCollector()
protos = coll.collect_for_message(Timestamp.DESCRIPTOR)
for p in protos:
    print(p.name)
```

## Run tests

```bash
cd python
make
```

Or directly with `uv`:

```bash
cd python
uv run pytest
```

## Example

See [`example/`](example/) for an end-to-end MCAP writer example. Run
`make example` from `python/` to build and run it; it produces
`python/example/output.mcap`.

See the repository [README.md](../README.md) for full cross-language
documentation.

