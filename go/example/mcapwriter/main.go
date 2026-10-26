// mcapwriter is a small example that writes an MCAP file using the
// filedesc-collector to supply protobuf FileDescriptorProtos and the Foxglove
// MCAP Go SDK to write the file.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/foxglove/mcap/go/mcap"
	"github.com/ytakio/filedesc-collector/go/filedesc"
	"github.com/ytakio/filedesc-collector/go/proto/foxglove"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Wrote output.mcap")
}

// channelConfig describes one channel to write: its topic, the protobuf message
// type used for its schema, and a populate function that builds each sample.
type channelConfig struct {
	topic     string
	schemaMsg proto.Message
	populate  func(t time.Time, i int) proto.Message
}

// channelConfigs is the table of channels produced by this example.
// Adding a new channel only requires appending a new entry here.
var channelConfigs = []channelConfig{
	{
		topic:     "/timestamp",
		schemaMsg: (*timestamppb.Timestamp)(nil),
		populate: func(t time.Time, i int) proto.Message {
			return &timestamppb.Timestamp{
				Seconds: t.Unix(),
				Nanos:   int32(t.Nanosecond()),
			}
		},
	},
	{
		topic:     "/sensor/dummy",
		schemaMsg: (*foxglove.SceneUpdate)(nil),
		populate: func(t time.Time, i int) proto.Message {
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

func run() (err error) {
	f, err := os.Create("output.mcap")
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close output file: %w", cerr)
		}
	}()

	w, err := mcap.NewWriter(f, &mcap.WriterOptions{
		Chunked: true,
	})
	if err != nil {
		return fmt.Errorf("create mcap writer: %w", err)
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close mcap writer: %w", cerr)
		}
	}()

	if err := w.WriteHeader(&mcap.Header{
		Profile: "protobuf",
		Library: "github.com/ytakio/filedesc-collector/mcapwriter example",
	}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Assign IDs dynamically instead of hardcoding them.
	var nextSchemaID uint16 = 1
	var nextChannelID uint16 = 0

	// Register schemas and channels from the channelConfigs table.
	channelIDs := make([]uint16, len(channelConfigs))
	for i, cfg := range channelConfigs {
		schema, err := createSchema(nextSchemaID, cfg.schemaMsg)
		if err != nil {
			return fmt.Errorf("create schema for %s: %w", cfg.topic, err)
		}
		if err := w.WriteSchema(schema); err != nil {
			return fmt.Errorf("write schema for %s: %w", cfg.topic, err)
		}
		nextSchemaID++

		channel := &mcap.Channel{
			ID:              nextChannelID,
			SchemaID:        schema.ID,
			Topic:           cfg.topic,
			MessageEncoding: "protobuf",
			Metadata:        map[string]string{"source": "mcapwriter example"},
		}
		if err := w.WriteChannel(channel); err != nil {
			return fmt.Errorf("write channel %s: %w", cfg.topic, err)
		}
		channelIDs[i] = nextChannelID
		nextChannelID++
	}

	// Write 100 samples at 1 Hz, starting from a fixed epoch.
	start := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	for i := range 100 {
		t := start.Add(time.Duration(i) * time.Second)
		nanos := uint64(t.UnixNano())

		for j, cfg := range channelConfigs {
			msg := cfg.populate(t, i)
			m, err := createMessage(
				channelIDs[j],
				uint32(i+1),
				nanos,
				msg,
			)
			if err != nil {
				return fmt.Errorf("create message %d on %s: %w", i, cfg.topic, err)
			}
			if err := w.WriteMessage(m); err != nil {
				return fmt.Errorf("write message %d on %s: %w", i, cfg.topic, err)
			}
		}
	}

	return nil
}

func createSchema(id uint16, msg proto.Message) (*mcap.Schema, error) {
	descriptor := msg.ProtoReflect()
	collector := filedesc.NewCollector()
	fdProtos, err := collector.CollectForMessage(descriptor)
	if err != nil {
		return nil, fmt.Errorf("collect file descriptors: %w", err)
	}

	fdSetBytes, err := proto.Marshal(&descriptorpb.FileDescriptorSet{
		File: fdProtos,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal file descriptor set: %w", err)
	}

	return &mcap.Schema{
		ID:       id,
		Name:     string(descriptor.Descriptor().FullName()),
		Encoding: "protobuf",
		Data:     fdSetBytes,
	}, nil
}

func createMessage(channelID uint16, sequence uint32, logTime uint64, msg proto.Message) (*mcap.Message, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	return &mcap.Message{
		ChannelID:   channelID,
		Sequence:    sequence,
		LogTime:     logTime,
		PublishTime: logTime,
		Data:        data,
	}, nil
}
