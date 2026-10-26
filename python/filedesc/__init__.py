"""Public API for the filedesc package.

This module re-exports :class:`FileDescriptorSetCollector`, the main entry
point for collecting protobuf FileDescriptorProtos including transitive
dependencies.
"""

from .collector import FileDescriptorSetCollector

__all__ = ["FileDescriptorSetCollector"]
