// sender publishes individual protobuf FileDescriptorProtos to NATS KV and
// serialized protobuf messages to NATS subjects.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ytakio/filedesc-collector/go/example/nats/common"
	"github.com/ytakio/filedesc-collector/go/filedesc"
	"google.golang.org/protobuf/proto"
)

const (
	// messageCount is the number of samples published on each channel.
	messageCount = 100
	// publishInterval is the time between samples.
	publishInterval = 100 * time.Millisecond
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("sender done")
}

func run() error {
	nc, err := nats.Connect(common.NatsURL)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("jetstream: %w", err)
	}

	schemasKV, err := common.GetOrCreateKV(js, common.SchemaBucket)
	if err != nil {
		return fmt.Errorf("schemas kv: %w", err)
	}

	subjectsKV, err := common.GetOrCreateKV(js, common.SubjectBucket)
	if err != nil {
		return fmt.Errorf("subjects kv: %w", err)
	}

	// Publish each channel's schema and subject-to-proto-path mapping first so
	// the receiver can resolve schemas before any data arrives.
	for _, cfg := range common.ChannelConfigs {
		subject := cfg.Subject
		md := cfg.MessageType.ProtoReflect().Descriptor()
		collector := filedesc.NewCollector()
		fdProtos, err := collector.CollectForMessage(cfg.MessageType.ProtoReflect())
		if err != nil {
			return fmt.Errorf("collect file descriptors for %s: %w", subject, err)
		}

		protoPath := string(md.ParentFile().Path())
		for _, fdProto := range fdProtos {
			path := fdProto.GetName()
			data, err := proto.Marshal(fdProto)
			if err != nil {
				return fmt.Errorf("marshal file descriptor %s for %s: %w", path, subject, err)
			}
			if _, err := schemasKV.Put(path, data); err != nil {
				return fmt.Errorf("put file descriptor %s for %s: %w", path, subject, err)
			}
			fmt.Printf("published schema %s -> %s (%d bytes)\n", subject, path, len(data))
		}

		info := common.SubjectInfo{
			Schema: common.SchemaRef{
				Path:     protoPath,
				FullName: string(md.FullName()),
			},
		}
		infoBytes, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("marshal subject info for %s: %w", subject, err)
		}
		if _, err := subjectsKV.Put(subject, infoBytes); err != nil {
			return fmt.Errorf("put subject info %s -> %s: %w", subject, infoBytes, err)
		}
		fmt.Printf("published subject info %s -> %s\n", subject, infoBytes)
	}

	// Publish serialized protobuf messages at a fixed rate.
	start := time.Now().UTC()
	for i := range messageCount {
		t := start.Add(time.Duration(i) * publishInterval)
		for _, cfg := range common.ChannelConfigs {
			msg := cfg.Populate(t, i)
			data, err := proto.Marshal(msg)
			if err != nil {
				return fmt.Errorf("marshal message for %s: %w", cfg.Subject, err)
			}
			if err := nc.Publish(cfg.Subject, data); err != nil {
				return fmt.Errorf("publish to %s: %w", cfg.Subject, err)
			}
		}
		time.Sleep(publishInterval)
	}
	return nil
}
