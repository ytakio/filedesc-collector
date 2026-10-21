"""Collect FileDescriptorProtos including transitive dependencies.

Package filedesc provides utilities for collecting FileDescriptorSet contents
from protobuf descriptors.
"""

from __future__ import annotations

from google.protobuf.descriptor import Descriptor, FileDescriptor
from google.protobuf.descriptor_pb2 import FileDescriptorProto


class FileDescriptorSetCollector:
    """Collects FileDescriptorProto messages including dependencies.

    A collector can be reused for multiple independent collections; its
    internal state is reset at the beginning of each collection.
    """

    def collect_for_message(self, descriptor: Descriptor) -> list[FileDescriptorProto]:
        """Return FileDescriptorProtos for the file containing ``descriptor`` and its deps."""
        if descriptor is None:
            raise ValueError("descriptor must not be None")
        return self.collect_for_descriptor(descriptor.file)

    def collect_for_descriptor(self, file_descriptor: FileDescriptor) -> list[FileDescriptorProto]:
        """Return FileDescriptorProtos for ``file_descriptor`` and its deps."""
        if file_descriptor is None:
            raise ValueError("file_descriptor must not be None")

        seen: set[str] = set()

        def collect_file(fd: FileDescriptor) -> list[FileDescriptorProto]:
            if fd is None:
                raise ValueError("file descriptor is None")

            path = fd.name
            if path in seen:
                return []
            seen.add(path)

            # Recursively collect dependencies first so the result is topologically ordered.
            result: list[FileDescriptorProto] = []
            for dep in fd.dependencies:
                result.extend(collect_file(dep))

            # Append this file's descriptor proto.
            file_proto = FileDescriptorProto()
            fd.CopyToProto(file_proto)
            result.append(file_proto)

            return result

        return collect_file(file_descriptor)

