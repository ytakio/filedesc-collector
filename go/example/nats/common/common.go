// Package common holds shared configuration and helpers for the NATS
// filedesc-collector example.
package common

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ytakio/filedesc-collector/go/example/proto/foxglove"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// NatsURL is the NATS client URL used by the example. It matches the
	// host port published by docker-compose.yml.
	NatsURL = "nats://localhost:4223"
	// SchemaBucket is the KV bucket name that stores FileDescriptorSet bytes
	// keyed by proto file path.
	SchemaBucket = "SCHEMAS"
	// SubjectBucket is the KV bucket name that stores schema metadata keyed by
	// NATS subject name.
	SubjectBucket = "SUBJECTS"
)

// SchemaRef identifies the protobuf schema used by messages published on a
// NATS subject.
type SchemaRef struct {
	// Path is the proto file path (FileDescriptorProto.name).
	Path string `json:"path"`
	// FullName is the fully-qualified descriptor name (DescriptorProto.full_name).
	FullName string `json:"full_name"`
}

// SubjectMapping is the value stored in the SUBJECTS KV bucket for each NATS
// subject.
type SubjectMapping struct {
	Schema SchemaRef `json:"schema"`
}

// ChannelConfig describes one channel produced/consumed by the example.
type ChannelConfig struct {
	Topic     string
	SchemaMsg proto.Message
	Populate  func(t time.Time, i int) proto.Message
}

// TopicToSubject converts an MCAP-style topic ("/sensor/dummy") into a NATS
// subject ("sensor.dummy"). A leading slash is stripped because NATS
// subjects cannot begin with a dot.
func TopicToSubject(topic string) string {
	return strings.ReplaceAll(strings.TrimPrefix(topic, "/"), "/", ".")
}

// SubjectToTopic converts a NATS subject ("sensor.dummy") into an MCAP-style
// topic ("/sensor/dummy").
func SubjectToTopic(subject string) string {
	return "/" + strings.ReplaceAll(subject, ".", "/")
}

// ProtoPath returns the proto file path for a message type.
func ProtoPath(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().ParentFile().Path())
}

// GetOrCreateKV opens an existing KV bucket or creates it when it does not
// exist yet.
func GetOrCreateKV(js nats.JetStreamContext, bucket string) (nats.KeyValue, error) {
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{Bucket: bucket})
	if err == nil {
		return kv, nil
	}

	var jsErr nats.JetStreamError
	if errors.As(err, &jsErr) {
		if apiErr := jsErr.APIError(); apiErr != nil && apiErr.ErrorCode == nats.JSErrCodeStreamNameInUse {
			return js.KeyValue(bucket)
		}
	}

	var apiErr *nats.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode == nats.JSErrCodeStreamNameInUse {
		return js.KeyValue(bucket)
	}
	return nil, fmt.Errorf("kv bucket %s: %w", bucket, err)
}

// ChannelConfigs is the table of channels used by both sender and receiver.
// It mirrors the mcapwriter example so the outputs can be compared.
var ChannelConfigs = []ChannelConfig{
	{
		Topic:     "/timestamp",
		SchemaMsg: (*timestamppb.Timestamp)(nil),
		Populate: func(t time.Time, i int) proto.Message {
			return &timestamppb.Timestamp{
				Seconds: t.Unix(),
				Nanos:   int32(t.Nanosecond()),
			}
		},
	},
	{
		Topic:     "/sensor/dummy",
		SchemaMsg: (*foxglove.SceneUpdate)(nil),
		Populate: func(t time.Time, i int) proto.Message {
			return &foxglove.SceneUpdate{
				Entities: []*foxglove.SceneEntity{
					{
						Id:      fmt.Sprintf("entity-%02d", i),
						FrameId: "base_link",
						Cubes: []*foxglove.CubePrimitive{
							{
								Pose: &foxglove.Pose{
									Position:    &foxglove.Vector3{X: float64(i), Y: 0, Z: 0},
									Orientation: &foxglove.Quaternion{X: 0, Y: 0, Z: 0, W: 1},
								},
								Size:  &foxglove.Vector3{X: 1, Y: 1, Z: 1},
								Color: &foxglove.Color{R: 1.0, G: 0.5, B: 0.0, A: 1.0},
							},
						},
					},
				},
			}
		},
	},
}
