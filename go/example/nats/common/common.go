// Package common holds shared configuration and helpers for the NATS
// filedesc-collector example.
package common

import (
	"fmt"
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

// SubjectInfo holds metadata for a NATS subject, including the schema used by
// messages published on that subject.
type SubjectInfo struct {
	Schema SchemaRef `json:"schema"`
}

// ChannelConfig describes one channel produced/consumed by the example.
type ChannelConfig struct {
	Subject string
	// MessageType is the prototype used to obtain the protobuf descriptor for
	// this channel. It holds a typed nil pointer such as (*T)(nil).
	MessageType proto.Message
	Populate    func(t time.Time, i int) proto.Message
}

// GetOrCreateKV opens an existing KV bucket or creates it when it does not
// exist yet.
func GetOrCreateKV(js nats.JetStreamContext, bucket string) (nats.KeyValue, error) {
	if kv, err := js.KeyValue(bucket); err == nil {
		return kv, nil
	}
	return js.CreateKeyValue(&nats.KeyValueConfig{Bucket: bucket})
}

// ChannelConfigs is the table of channels used by both sender and receiver.
// It mirrors the mcapwriter example so the outputs can be compared.
var ChannelConfigs = []ChannelConfig{
	{
		Subject:     "timestamp",
		MessageType: (*timestamppb.Timestamp)(nil),
		Populate: func(t time.Time, i int) proto.Message {
			return &timestamppb.Timestamp{
				Seconds: t.Unix(),
				Nanos:   int32(t.Nanosecond()),
			}
		},
	},
	{
		Subject:     "sensor.dummy",
		MessageType: (*foxglove.SceneUpdate)(nil),
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
