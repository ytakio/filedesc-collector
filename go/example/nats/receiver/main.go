// receiver subscribes to fixed NATS subjects, collects individual protobuf
// FileDescriptorProtos from NATS KV into a FileDescriptorSet, and writes the
// raw messages to an MCAP file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/foxglove/mcap/go/mcap"
	"github.com/nats-io/nats.go"
	"github.com/ytakio/filedesc-collector/go/example/nats/common"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	outputPath     = "output.mcap"
	receiveTimeout = 15 * time.Second
	watchTimeout   = 30 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("receiver done")
}

func run() (err error) {
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

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w, err := mcap.NewWriter(f, &mcap.WriterOptions{Chunked: true})
	if err != nil {
		return fmt.Errorf("create mcap writer: %w", err)
	}
	defer func() { _ = w.Close() }()

	if err := w.WriteHeader(&mcap.Header{
		Profile: "protobuf",
		Library: "github.com/ytakio/filedesc-collector/nats receiver example",
	}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Resolve each subject's info and schema chain, register the corresponding
	// MCAP channel, and subscribe one subject at a time.
	watchCtx, cancel := context.WithTimeout(context.Background(), watchTimeout)
	defer cancel()

	channelIDs := make(map[string]uint16)
	var nextSchema uint16 = 1
	var nextChannel uint16

	ctx, cancel := context.WithTimeout(context.Background(), receiveTimeout)
	defer cancel()

	msgChan := make(chan *nats.Msg, 256)
	for _, cfg := range common.ChannelConfigs {
		entry, err := waitForKey(watchCtx, subjectsKV, cfg.Subject)
		if err != nil {
			return fmt.Errorf("subject %s: %w", cfg.Subject, err)
		}
		var info common.SubjectInfo
		if err := json.Unmarshal(entry.Value(), &info); err != nil {
			return fmt.Errorf("unmarshal %s: %w", cfg.Subject, err)
		}
		fmt.Printf("received subject info %s -> %s\n", cfg.Subject, info.Schema.Path)

		fds, err := resolveFileDescriptorSet(watchCtx, schemasKV, info.Schema.Path, cfg.Subject)
		if err != nil {
			return fmt.Errorf("resolve schema for %s: %w", cfg.Subject, err)
		}
		data, err := proto.Marshal(fds)
		if err != nil {
			return fmt.Errorf("marshal fds for %s: %w", cfg.Subject, err)
		}
		if err := w.WriteSchema(&mcap.Schema{
			ID:       nextSchema,
			Name:     info.Schema.FullName,
			Encoding: "protobuf",
			Data:     data,
		}); err != nil {
			return fmt.Errorf("write schema for %s: %w", cfg.Subject, err)
		}
		if err := w.WriteChannel(&mcap.Channel{
			ID:              nextChannel,
			SchemaID:        nextSchema,
			Topic:           cfg.Subject,
			MessageEncoding: "protobuf",
			Metadata:        map[string]string{"source": "nats receiver example"},
		}); err != nil {
			return fmt.Errorf("write channel for %s: %w", cfg.Subject, err)
		}
		channelIDs[cfg.Subject] = nextChannel
		nextSchema++
		nextChannel++
		fmt.Printf("registered schema %s -> %s\n", cfg.Subject, info.Schema.Path)

		if _, err := nc.Subscribe(cfg.Subject, func(m *nats.Msg) {
			select {
			case msgChan <- m:
			case <-ctx.Done():
			}
		}); err != nil {
			return fmt.Errorf("subscribe to %s: %w", cfg.Subject, err)
		}
	}

	// Receive messages until timeout.
	done := make(chan struct{})
	go func() {
		defer close(done)
		var sequence uint32
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-msgChan:
				sequence++
				logTime := uint64(time.Now().UnixNano())
				if err := w.WriteMessage(&mcap.Message{
					ChannelID:   channelIDs[m.Subject],
					Sequence:    sequence,
					LogTime:     logTime,
					PublishTime: logTime,
					Data:        m.Data,
				}); err != nil {
					fmt.Fprintf(os.Stderr, "write message: %v\n", err)
					return
				}
			}
		}
	}()

	<-done

	return nil
}

// waitForKey blocks until the given KV key is set to a value.
func waitForKey(ctx context.Context, kv nats.KeyValue, key string) (nats.KeyValueEntry, error) {
	watcher, err := kv.Watch(key, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("watch %s: %w", key, err)
	}
	defer func() { _ = watcher.Stop() }()

	for entry := range watcher.Updates() {
		if entry == nil {
			continue
		}
		if entry.Operation() == nats.KeyValuePut {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("watcher for %s closed", key)
}

// resolveFileDescriptorSet waits for the file descriptor at protoPath (and its
// transitive dependencies) from KV and returns a FileDescriptorSet in
// dependency order.
func resolveFileDescriptorSet(ctx context.Context, kv nats.KeyValue, protoPath string, subject string) (*descriptorpb.FileDescriptorSet, error) {
	var files []*descriptorpb.FileDescriptorProto
	seen := make(map[string]struct{})

	var resolve func(string) error
	resolve = func(path string) error {
		if _, ok := seen[path]; ok {
			return nil
		}
		seen[path] = struct{}{}

		entry, err := waitForKey(ctx, kv, path)
		if err != nil {
			return fmt.Errorf("schema %s: %w", path, err)
		}

		var fd descriptorpb.FileDescriptorProto
		if err := proto.Unmarshal(entry.Value(), &fd); err != nil {
			return fmt.Errorf("unmarshal %s: %w", path, err)
		}
		fmt.Printf("received schema %s -> %s (%d bytes)\n", subject, path, len(entry.Value()))

		for _, dep := range fd.Dependency {
			if err := resolve(dep); err != nil {
				return err
			}
		}
		files = append(files, &fd)
		return nil
	}

	if err := resolve(protoPath); err != nil {
		return nil, err
	}
	return &descriptorpb.FileDescriptorSet{File: files}, nil
}
