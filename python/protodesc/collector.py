"""Collect FileDescriptorProtos including transitive dependencies."""

from __future__ import annotations

from google.protobuf.descriptor import Descriptor, FileDescriptor
from google.protobuf.descriptor_pb2 import FileDescriptorProto


class FileDescriptorSetCollector:
    """Collects FileDescriptorProto messages including dependencies."""

    def __init__(self) -> None:
        self._seen: set[str] = set()

    def collect_for_message(self, descriptor: Descriptor) -> list[FileDescriptorProto]:
        """Return FileDescriptorProtos for the file containing ``descriptor`` and its deps."""
        if descriptor is None:
            raise ValueError("descriptor must not be None")
        return self._collect(descriptor.file)

    def collect_for_descriptor(self, file_descriptor: FileDescriptor) -> list[FileDescriptorProto]:
        """Return FileDescriptorProtos for ``file_descriptor`` and its deps."""
        if file_descriptor is None:
            raise ValueError("file_descriptor must not be None")
        return self._collect(file_descriptor)

    def _collect(self, file_descriptor: FileDescriptor) -> list[FileDescriptorProto]:
        result: list[FileDescriptorProto] = []
        self._collect_file(file_descriptor, result)
        return result

    def _collect_file(
        self, file_descriptor: FileDescriptor, result: list[FileDescriptorProto]
    ) -> None:
        if file_descriptor is None:
            return
        path = file_descriptor.name
        if path in self._seen:
            return
        self._seen.add(path)

        # Recursively collect dependencies first.
        for dep in file_descriptor.dependencies:
            self._collect_file(dep, result)

        # Append this file's descriptor proto.
        file_proto = FileDescriptorProto()
        file_descriptor.CopyToProto(file_proto)
        result.append(file_proto)


def collect_for_message(descriptor: Descriptor) -> list[FileDescriptorProto]:
    """Convenience function to collect FileDescriptorProtos for a message type."""
    return FileDescriptorSetCollector().collect_for_message(descriptor)


def collect_for_descriptor(file_descriptor: FileDescriptor) -> list[FileDescriptorProto]:
    """Convenience function to collect FileDescriptorProtos for a file descriptor."""
    return FileDescriptorSetCollector().collect_for_descriptor(file_descriptor)

