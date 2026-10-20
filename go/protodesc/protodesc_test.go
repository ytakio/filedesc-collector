package protodesc

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCollectForMessage(t *testing.T) {
	protos, err := CollectForMessage((*timestamppb.Timestamp)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("CollectForMessage failed: %v", err)
	}
	if len(protos) == 0 {
		t.Fatal("expected at least one FileDescriptorProto")
	}
	// The last proto should be google/protobuf/timestamp.proto.
	last := protos[len(protos)-1]
	if last.GetName() != "google/protobuf/timestamp.proto" {
		t.Fatalf("expected timestamp.proto, got %s", last.GetName())
	}
}

func TestCollectForMessageWithDependency(t *testing.T) {
	// anypb.Any depends on google/protobuf/any.proto only (no other well-known deps).
	protos, err := CollectForMessage((*anypb.Any)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("CollectForMessage failed: %v", err)
	}
	found := false
	for _, f := range protos {
		if f.GetName() == "google/protobuf/any.proto" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected google/protobuf/any.proto in collected protos")
	}
}

func TestCollectForDescriptor(t *testing.T) {
	protos, err := CollectForDescriptor((&descriptorpb.FileDescriptorSet{}).ProtoReflect().Descriptor().ParentFile())
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
	if len(first) != 1 {
		t.Fatalf("expected first collect to contain 1 file, got %d", len(first))
	}

	second, err := c.CollectForMessage((*anypb.Any)(nil).ProtoReflect())
	if err != nil {
		t.Fatalf("second collect failed: %v", err)
	}
	if len(second) != 1 {
		t.Fatalf("expected second collect to contain 1 file (any.proto), got %d", len(second))
	}
	if second[0].GetName() != "google/protobuf/any.proto" {
		t.Fatalf("expected second collect to be any.proto, got %s", second[0].GetName())
	}
}

func TestCollectForNil(t *testing.T) {
	var msg protoreflect.Message
	if _, err := CollectForMessage(msg); err == nil {
		t.Fatal("expected error for nil message")
	}
}
