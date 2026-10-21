package filedesc

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func fileNames(protos []*descriptorpb.FileDescriptorProto) []string {
	names := make([]string, len(protos))
	for i, p := range protos {
		names[i] = p.GetName()
	}
	return names
}

func TestCollectForMessage(t *testing.T) {
	protos, err := NewCollector().CollectForMessage((*timestamppb.Timestamp)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("CollectForMessage failed: %v", err)
	}
	if len(protos) == 0 {
		t.Fatal("expected at least one FileDescriptorProto")
	}
	if got, want := protos[len(protos)-1].GetName(), "google/protobuf/timestamp.proto"; got != want {
		t.Fatalf("last file = %q, want %q", got, want)
	}
}

func TestCollectForMessageWithDependency(t *testing.T) {
	// anypb.Any depends on google/protobuf/any.proto only (no other well-known deps).
	protos, err := NewCollector().CollectForMessage((*anypb.Any)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("CollectForMessage failed: %v", err)
	}
	if !slices.Contains(fileNames(protos), "google/protobuf/any.proto") {
		t.Fatal("expected google/protobuf/any.proto in collected protos")
	}
}

func TestCollectForDescriptor(t *testing.T) {
	protos, err := NewCollector().CollectForDescriptor((&descriptorpb.FileDescriptorSet{}).ProtoReflect().Descriptor().ParentFile())
	if err != nil {
		t.Fatalf("CollectForDescriptor failed: %v", err)
	}
	if len(protos) == 0 {
		t.Fatal("expected at least one FileDescriptorProto")
	}
}

func TestCollectorDoesNotDuplicate(t *testing.T) {
	c := NewCollector()
	first, err := c.CollectForMessage((*timestamppb.Timestamp)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("first collect failed: %v", err)
	}
	if got, want := len(first), 1; got != want {
		t.Fatalf("first collect length = %d, want %d", got, want)
	}

	second, err := c.CollectForMessage((*anypb.Any)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("second collect failed: %v", err)
	}
	if got, want := len(second), 1; got != want {
		t.Fatalf("second collect length = %d, want %d", got, want)
	}
	if got, want := second[0].GetName(), "google/protobuf/any.proto"; got != want {
		t.Fatalf("second collect file = %q, want %q", got, want)
	}
}

func TestCollectForNil(t *testing.T) {
	var msg protoreflect.Message
	if _, err := NewCollector().CollectForMessage(msg); err == nil {
		t.Fatal("expected error for nil message")
	}
}
