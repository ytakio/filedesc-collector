// receiver subscribes to fixed NATS subjects, collects individual protobuf
// FileDescriptorProtos from NATS KV into a FileDescriptorSet, parses messages
// with dynamic descriptors, and writes them to an MCAP file.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
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

	subjectsKV, err := common.GetOrCreateKV(js, common.SubjectBucket)
	if err != nil {
		return fmt.Errorf("subjects kv: %w", err)
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
	var subs []*nats.Subscription
	for _, cfg := range common.ChannelConfigs {
		subject := common.TopicToSubject(cfg.Topic)
		sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
			select {
			case msgChan <- &natsMsg{msg: m, subject: m.Subject}:
			case <-ctx.Done():
			}
		})
		if err != nil {
			return fmt.Errorf("subscribe to %s: %w", subject, err)
		}
		subs = append(subs, sub)
	}

	state := &writeState{
		w:               w,
		schemasKV:       schemasKV,
		subjectsKV:      subjectsKV,
		schemas:         make(map[string]*schema),
		subjectMappings: make(map[string]*subjectMapping),
		schemaIDs:       make(map[string]uint16),
		channelIDs:      make(map[string]uint16),
		nextSchema:      1, // MCAP schema IDs must be non-zero.
		nextChannel:     0,
	}

	// Resolve every target subject's schema mapping and protobuf schema before
	// subscribing so message handling uses only pre-resolved data.
	if err := state.prefetchSubjectMappings(); err != nil {
		return fmt.Errorf("prefetch subject mappings: %w", err)
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
	msg     *nats.Msg
	subject string
}

type schema struct {
	schemaBytes []byte
	md          protoreflect.MessageDescriptor
}

type subjectMapping struct {
	protoPath string
	fullName  protoreflect.FullName
}

type writeState struct {
	w               *mcap.Writer
	schemasKV       nats.KeyValue
	subjectsKV      nats.KeyValue
	subjectMappings map[string]*subjectMapping
	schemas         map[string]*schema
	schemaIDs       map[string]uint16
	channelIDs      map[string]uint16
	nextSchema      uint16
	nextChannel     uint16
}

func (s *writeState) prefetchSubjectMappings() error {
	for _, cfg := range common.ChannelConfigs {
		subject := common.TopicToSubject(cfg.Topic)
		entry, err := s.waitForSubjectMapping(subject)
		if err != nil {
			return fmt.Errorf("get subject mapping %s: %w", subject, err)
		}

		var mapping common.SubjectMapping
		if err := json.Unmarshal(entry.Value(), &mapping); err != nil {
			return fmt.Errorf("unmarshal subject mapping for %s: %w", subject, err)
		}

		fullName := protoreflect.FullName(mapping.Schema.FullName)
		s.subjectMappings[subject] = &subjectMapping{
			protoPath: mapping.Schema.Path,
			fullName:  fullName,
		}

		if _, _, err := s.resolveSchema(mapping.Schema.Path, fullName); err != nil {
			return fmt.Errorf("prefetch schema %s for %s: %w", mapping.Schema.Path, subject, err)
		}
		fmt.Printf("prefetched schema for %s -> %s (%s)\n", subject, mapping.Schema.Path, fullName)
	}
	return nil
}

// waitForSubjectMapping retries reading the SUBJECTS KV entry until the sender
// has published it or a generous timeout expires.
func (s *writeState) waitForSubjectMapping(subject string) (nats.KeyValueEntry, error) {
	const (
		maxAttempts = 60
		delay       = 500 * time.Millisecond
	)
	for range maxAttempts {
		entry, err := s.subjectsKV.Get(subject)
		if err == nil {
			return entry, nil
		}
		if !errors.Is(err, nats.ErrKeyNotFound) {
			return nil, err
		}
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("subject mapping %s not found after %v", subject, maxAttempts*delay)
}

func (s *writeState) writeMessage(m *natsMsg, sequence uint32, logTime uint64) error {
	mapping, ok := s.subjectMappings[m.subject]
	if !ok {
		return fmt.Errorf("no subject mapping for subject %s", m.subject)
	}

	schemaKey := string(mapping.fullName)
	schema, ok := s.schemas[schemaKey]
	if !ok {
		return fmt.Errorf("no schema for %s", mapping.fullName)
	}

	// Demonstrate dynamic parsing with the schema prepared before subscribe.
	dynMsg := dynamicpb.NewMessage(schema.md)
	if err := proto.Unmarshal(m.msg.Data, dynMsg); err != nil {
		return fmt.Errorf("unmarshal dynamic message %s: %w", mapping.fullName, err)
	}

	topic := common.SubjectToTopic(m.subject)
	schemaID, err := s.ensureSchema(topic, mapping.fullName, schema.schemaBytes)
	if err != nil {
		return fmt.Errorf("ensure schema for %s: %w", topic, err)
	}

	channelID, err := s.ensureChannel(topic, schemaID)
	if err != nil {
		return fmt.Errorf("ensure channel for %s: %w", topic, err)
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

func (s *writeState) resolveSchema(protoPath string, fullName protoreflect.FullName) ([]byte, protoreflect.MessageDescriptor, error) {
	schemaKey := string(fullName)
	if schema, ok := s.schemas[schemaKey]; ok {
		return schema.schemaBytes, schema.md, nil
	}

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

	s.schemas[schemaKey] = &schema{schemaBytes: schemaBytes, md: md}
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
