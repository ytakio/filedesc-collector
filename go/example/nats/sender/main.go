// sender publishes protobuf FileDescriptorSets to NATS KV and serialized
// protobuf messages to NATS subjects.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ytakio/filedesc-collector/go/example/nats/common"
	"github.com/ytakio/filedesc-collector/go/filedesc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
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

	topicsKV, err := common.GetOrCreateKV(js, common.TopicBucket)
	if err != nil {
		return fmt.Errorf("topics kv: %w", err)
	}

	// Publish each channel's schema and topic-to-proto-path mapping first so
	// the receiver can resolve schemas before any data arrives.
	for _, cfg := range common.ChannelConfigs {
		descriptor := cfg.SchemaMsg.ProtoReflect()
		collector := filedesc.NewCollector()
		fdProtos, err := collector.CollectForMessage(descriptor)
		if err != nil {
			return fmt.Errorf("collect file descriptors for %s: %w", cfg.Topic, err)
		}

		fdSetBytes, err := proto.Marshal(&descriptorpb.FileDescriptorSet{File: fdProtos})
		if err != nil {
			return fmt.Errorf("marshal file descriptor set for %s: %w", cfg.Topic, err)
		}

		protoPath := common.ProtoPath(cfg.SchemaMsg)
		if _, err := schemasKV.Put(protoPath, fdSetBytes); err != nil {
			return fmt.Errorf("put schema %s for %s: %w", protoPath, cfg.Topic, err)
		}
		fmt.Printf("published schema %s -> %s (%d bytes)\n", cfg.Topic, protoPath, len(fdSetBytes))

		msgName := string(descriptor.Descriptor().FullName())
		topicValue := fmt.Sprintf("%s|%s", protoPath, msgName)
		if _, err := topicsKV.Put(cfg.Topic, []byte(topicValue)); err != nil {
			return fmt.Errorf("put topic mapping %s -> %s: %w", cfg.Topic, topicValue, err)
		}
		fmt.Printf("published topic mapping %s -> %s\n", cfg.Topic, topicValue)
	}

	// Publish serialized protobuf messages at a fixed rate.
	start := time.Now().UTC()
	for i := range messageCount {
		t := start.Add(time.Duration(i) * publishInterval)
		for _, cfg := range common.ChannelConfigs {
			msg := cfg.Populate(t, i)
			data, err := proto.Marshal(msg)
			if err != nil {
				return fmt.Errorf("marshal message for %s: %w", cfg.Topic, err)
			}
			subject := common.TopicToSubject(cfg.Topic)
			if err := nc.Publish(subject, data); err != nil {
				return fmt.Errorf("publish to %s: %w", subject, err)
			}
		}
		time.Sleep(publishInterval)
	}

	fmt.Printf("published %d messages on each of %d channels\n", messageCount, len(common.ChannelConfigs))
	return nil
}
