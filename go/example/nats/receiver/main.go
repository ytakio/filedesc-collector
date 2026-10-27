// receiver subscribes to fixed NATS subjects, collects individual protobuf
// FileDescriptorProtos from NATS KV into a FileDescriptorSet, parses messages
// with dynamic descriptors, and writes them to an MCAP file.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/foxglove/mcap/go/mcap"
	"github.com/nats-io/nats.go"
	"github.com/ytakio/filedesc-collector/go/example/nats/common"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

const receiveTimeout = 15 * time.Second

func main() {
	outputPath := flag.String("output", "output.mcap", "output MCAP file path")
	flag.Parse()

	if err := run(*outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("receiver done")
}

func run(outputPath string) (err error) {
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

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close output file: %w", cerr)
		}
	}()

	w, err := mcap.NewWriter(f, &mcap.WriterOptions{Chunked: true})
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
		Library: "github.com/ytakio/filedesc-collector/nats receiver example",
	}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), receiveTimeout)
	defer cancel()

	msgChan := make(chan *natsMsg, 256)
	subjectToTopic := make(map[string]string, len(common.ChannelConfigs))
	var subs []*nats.Subscription
	for _, cfg := range common.ChannelConfigs {
		subject := common.TopicToSubject(cfg.Topic)
		subjectToTopic[subject] = cfg.Topic
		sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
			select {
			case msgChan <- &natsMsg{msg: m, topic: subjectToTopic[m.Subject]}:
			case <-ctx.Done():
			}
		})
		if err != nil {
			return fmt.Errorf("subscribe to %s: %w", subject, err)
		}
		subs = append(subs, sub)
	}

	state := &writeState{
		w:           w,
		schemasKV:   schemasKV,
		topicsKV:    topicsKV,
		schemaIDs:   make(map[string]uint16),
		channelIDs:  make(map[string]uint16),
		nextSchema:  1, // MCAP schema IDs must be non-zero.
		nextChannel: 0,
	}

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
				if err := state.writeMessage(m, sequence, logTime); err != nil {
					fmt.Fprintf(os.Stderr, "write message: %v\n", err)
					return
				}
			}
		}
	}()

	<-ctx.Done()
	for _, sub := range subs {
		_ = sub.Unsubscribe()
	}
	<-done

	return nil
}

type natsMsg struct {
	msg   *nats.Msg
	topic string
}

type writeState struct {
	w           *mcap.Writer
	schemasKV   nats.KeyValue
	topicsKV    nats.KeyValue
	schemaIDs   map[string]uint16
	channelIDs  map[string]uint16
	nextSchema  uint16
	nextChannel uint16
}

func (s *writeState) writeMessage(m *natsMsg, sequence uint32, logTime uint64) error {
	protoPath, fullName, err := s.resolveProtoPath(m.topic)
	if err != nil {
		return fmt.Errorf("resolve proto path for %s: %w", m.topic, err)
	}

	schemaBytes, md, err := s.resolveSchema(protoPath, fullName)
	if err != nil {
		return fmt.Errorf("resolve schema %s: %w", protoPath, err)
	}

	// Demonstrate dynamic parsing with the schema resolved from NATS KV.
	dynMsg := dynamicpb.NewMessage(md)
	if err := proto.Unmarshal(m.msg.Data, dynMsg); err != nil {
		return fmt.Errorf("unmarshal dynamic message %s: %w", fullName, err)
	}

	schemaID, err := s.ensureSchema(m.topic, fullName, schemaBytes)
	if err != nil {
		return fmt.Errorf("ensure schema for %s: %w", m.topic, err)
	}

	channelID, err := s.ensureChannel(m.topic, schemaID)
	if err != nil {
		return fmt.Errorf("ensure channel for %s: %w", m.topic, err)
	}

	if err := s.w.WriteMessage(&mcap.Message{
		ChannelID:   channelID,
		Sequence:    sequence,
		LogTime:     logTime,
		PublishTime: logTime,
		Data:        m.msg.Data,
	}); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (s *writeState) resolveProtoPath(topic string) (string, protoreflect.FullName, error) {
	entry, err := s.topicsKV.Get(topic)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(entry.Value()), "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid topic value: %s", entry.Value())
	}
	return parts[0], protoreflect.FullName(parts[1]), nil
}

func (s *writeState) resolveSchema(protoPath string, fullName protoreflect.FullName) ([]byte, protoreflect.MessageDescriptor, error) {
	fdSet, err := s.collectFileDescriptorSet(protoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("collect file descriptor set: %w", err)
	}
	schemaBytes, err := proto.Marshal(fdSet)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal file descriptor set: %w", err)
	}

	files, err := protodesc.NewFiles(fdSet)
	if err != nil {
		return nil, nil, fmt.Errorf("build file registry: %w", err)
	}

	d, err := files.FindDescriptorByName(fullName)
	if err != nil {
		return nil, nil, fmt.Errorf("find descriptor %s: %w", fullName, err)
	}

	md, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, nil, fmt.Errorf("descriptor %s is not a message", fullName)
	}

	return schemaBytes, md, nil
}

func (s *writeState) collectFileDescriptorSet(protoPath string) (*descriptorpb.FileDescriptorSet, error) {
	seen := make(map[string]struct{})
	var collect func(string) ([]*descriptorpb.FileDescriptorProto, error)
	collect = func(path string) ([]*descriptorpb.FileDescriptorProto, error) {
		if _, ok := seen[path]; ok {
			return nil, nil
		}
		seen[path] = struct{}{}

		entry, err := s.schemasKV.Get(path)
		if err != nil {
			return nil, fmt.Errorf("get file descriptor %s: %w", path, err)
		}
		var fdProto descriptorpb.FileDescriptorProto
		if err := proto.Unmarshal(entry.Value(), &fdProto); err != nil {
			return nil, fmt.Errorf("unmarshal file descriptor %s: %w", path, err)
		}

		var result []*descriptorpb.FileDescriptorProto
		for _, dep := range fdProto.Dependency {
			depProtos, err := collect(dep)
			if err != nil {
				return nil, err
			}
			result = append(result, depProtos...)
		}
		return append(result, &fdProto), nil
	}

	fdProtos, err := collect(protoPath)
	if err != nil {
		return nil, err
	}
	return &descriptorpb.FileDescriptorSet{File: fdProtos}, nil
}

func (s *writeState) ensureSchema(topic string, fullName protoreflect.FullName, schemaBytes []byte) (uint16, error) {
	if id, ok := s.schemaIDs[topic]; ok {
		return id, nil
	}
	id := s.nextSchema
	s.nextSchema++
	s.schemaIDs[topic] = id

	if err := s.w.WriteSchema(&mcap.Schema{
		ID:       id,
		Name:     string(fullName),
		Encoding: "protobuf",
		Data:     schemaBytes,
	}); err != nil {
		return 0, fmt.Errorf("write schema: %w", err)
	}
	return id, nil
}

func (s *writeState) ensureChannel(topic string, schemaID uint16) (uint16, error) {
	if id, ok := s.channelIDs[topic]; ok {
		return id, nil
	}
	id := s.nextChannel
	s.nextChannel++
	s.channelIDs[topic] = id

	if err := s.w.WriteChannel(&mcap.Channel{
		ID:              id,
		SchemaID:        schemaID,
		Topic:           topic,
		MessageEncoding: "protobuf",
		Metadata:        map[string]string{"source": "nats receiver example"},
	}); err != nil {
		return 0, fmt.Errorf("write channel: %w", err)
	}
	return id, nil
}
