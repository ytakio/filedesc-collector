// Package filedesc provides utilities for working with protobuf descriptors, including collecting FileDescriptorSets for messages and file descriptors.
package filedesc

import (
	"errors"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// FileDescriptorSetCollector collects FileDescriptorProtos including transitive dependencies.
// A collector can be reused for multiple independent collections; its internal state is
// reset at the beginning of each collection.
type FileDescriptorSetCollector struct{}

// NewCollector creates a new FileDescriptorSetCollector.
func NewCollector() *FileDescriptorSetCollector {
	return &FileDescriptorSetCollector{}
}

// CollectForDescriptor returns a list of FileDescriptorProto messages that include the
// given file descriptor and all of its transitive dependencies. Dependencies appear
// before the files that depend on them.
func (c *FileDescriptorSetCollector) CollectForDescriptor(fd protoreflect.FileDescriptor) ([]*descriptorpb.FileDescriptorProto, error) {
	if fd == nil {
		return nil, errors.New("protoreflect.FileDescriptor must not be nil")
	}
	state := &collectState{seen: make(map[string]struct{})}
	return state.collectFile(fd)
}

// CollectForMessage is a convenience method that returns FileDescriptorProtos for the
// file that defines the given message type.
func (c *FileDescriptorSetCollector) CollectForMessage(msg protoreflect.Message) ([]*descriptorpb.FileDescriptorProto, error) {
	if msg == nil {
		return nil, errors.New("protoreflect.Message must not be nil")
	}
	return c.CollectForDescriptor(msg.Descriptor().ParentFile())
}

// collectState holds the mutable state for a single collection.
type collectState struct {
	seen map[string]struct{}
}

// collectFile recursively collects the given file descriptor and its transitive
// dependencies. Each file is emitted only once, and dependencies are emitted before
// the file that imports them.
func (s *collectState) collectFile(fd protoreflect.FileDescriptor) ([]*descriptorpb.FileDescriptorProto, error) {
	if fd == nil {
		return nil, errors.New("file descriptor is nil")
	}

	path := fd.Path()
	if _, ok := s.seen[path]; ok {
		return nil, nil
	}
	s.seen[path] = struct{}{}

	// Recursively collect dependencies first so the result is topologically ordered.
	var result []*descriptorpb.FileDescriptorProto
	deps := fd.Imports()
	for i := range deps.Len() {
		depProtos, err := s.collectFile(deps.Get(i))
		if err != nil {
			return nil, err
		}
		result = append(result, depProtos...)
	}

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
