// Package protodesc provides utilities for working with protobuf descriptors, including collecting FileDescriptorSets for messages and file descriptors.
package protodesc

import (
	"errors"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// FileDescriptorSetCollector collects FileDescriptorSets including dependencies.
type FileDescriptorSetCollector struct {
	seen map[string]struct{}
}

// NewCollector creates a new FileDescriptorSetCollector.
func NewCollector() *FileDescriptorSetCollector {
	return &FileDescriptorSetCollector{
		seen: make(map[string]struct{}),
	}
}

// CollectForMessage returns a list of FileDescriptorProto messages that include the
// file containing the given message type and all of its transitive dependencies.
func (c *FileDescriptorSetCollector) CollectForMessage(msg protoreflect.Message) ([]*descriptorpb.FileDescriptorProto, error) {
	if msg == nil {
		return nil, errors.New("protoreflect.Message must not be nil")
	}
	return c.CollectForDescriptor(msg.Descriptor().ParentFile())
}

// CollectForDescriptor returns a list of FileDescriptorProto messages that include the
// given file descriptor and all of its transitive dependencies.
func (c *FileDescriptorSetCollector) CollectForDescriptor(fd protoreflect.FileDescriptor) ([]*descriptorpb.FileDescriptorProto, error) {
	if fd == nil {
		return nil, errors.New("protoreflect.FileDescriptor must not be nil")
	}
	return c.collectFile(fd)
}

func (c *FileDescriptorSetCollector) collectFile(fd protoreflect.FileDescriptor) ([]*descriptorpb.FileDescriptorProto, error) {
	if fd == nil {
		return nil, errors.New("file descriptor is nil")
	}
	path := fd.Path()
	if _, ok := c.seen[path]; ok {
		return nil, nil
	}
	c.seen[path] = struct{}{}

	// Recursively collect dependencies first.
	var result []*descriptorpb.FileDescriptorProto
	deps := fd.Imports()
	for i := range deps.Len() {
		depProtos, err := c.collectFile(deps.Get(i))
		if err != nil {
			return nil, err
		}
		result = append(result, depProtos...)
	}

	// Append this file's descriptor proto.
	result = append(result, protodesc.ToFileDescriptorProto(fd))
	return result, nil
}

// CollectForMessage is a convenience function that creates a new collector and
// returns FileDescriptorProtos for the given message type.
func CollectForMessage(msg protoreflect.Message) ([]*descriptorpb.FileDescriptorProto, error) {
	return NewCollector().CollectForMessage(msg)
}

// CollectForDescriptor is a convenience function that creates a new collector and
// returns FileDescriptorProtos for the given file descriptor.
func CollectForDescriptor(fd protoreflect.FileDescriptor) ([]*descriptorpb.FileDescriptorProto, error) {
	return NewCollector().CollectForDescriptor(fd)
}
