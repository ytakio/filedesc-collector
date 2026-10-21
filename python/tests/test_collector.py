from google.protobuf.descriptor_pb2 import FileDescriptorProto, FileDescriptorSet
from google.protobuf.duration_pb2 import Duration
from google.protobuf.timestamp_pb2 import Timestamp
import pytest

from filedesc import (
    FileDescriptorSetCollector,
    collect_for_descriptor,
    collect_for_message,
)


def _file_names(protos: list[FileDescriptorProto]) -> list[str]:
    return [f.name for f in protos]


def test_collect_for_message():
    protos = collect_for_message(Timestamp.DESCRIPTOR)
    assert len(protos) > 0
    assert _file_names(protos)[-1] == "google/protobuf/timestamp.proto"


def test_collect_for_message_with_dependency():
    # Timestamp has no external deps, so just verify the target file is present.
    protos = collect_for_message(Duration.DESCRIPTOR)
    assert "google/protobuf/duration.proto" in _file_names(protos)


def test_collect_for_descriptor():
    protos = collect_for_descriptor(FileDescriptorSet.DESCRIPTOR.file)
    assert len(protos) > 0


def test_collector_does_not_duplicate():
    collector = FileDescriptorSetCollector()
    first = collector.collect_for_message(Timestamp.DESCRIPTOR)
    assert len(first) == 1
    assert first[0].name == "google/protobuf/timestamp.proto"

    second = collector.collect_for_message(Duration.DESCRIPTOR)
    assert len(second) == 1
    assert second[0].name == "google/protobuf/duration.proto"


def test_collect_for_none_message():
    with pytest.raises(ValueError):
        collect_for_message(None)


def test_collect_for_none_descriptor():
    with pytest.raises(ValueError):
        collect_for_descriptor(None)

