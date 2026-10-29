.PHONY: all example up down sender receiver clean protoc

# This Makefile is at go/example/nats/. The Go module root is two levels up
# and the go/example directory (which owns protobuf generation) is one level up.
GO_MODULE_ROOT := ../..
EXAMPLE_DIR := ..

all: example

protoc:
	$(MAKE) -C $(EXAMPLE_DIR) protoc

up:
	docker-compose up -d

down:
	docker-compose down

sender: protoc
	cd $(GO_MODULE_ROOT)/example/nats && go run ./sender

receiver: protoc
	cd $(GO_MODULE_ROOT)/example/nats && go run ./receiver

# Run the full example: start NATS, run the receiver in the background so it
# can wait for schemas via KV watch, then run the sender, wait for the receiver
# to finish, and tear NATS down.
example: protoc up
	cd $(GO_MODULE_ROOT)/example/nats && \
	go run ./receiver & \
	RECV_PID=$$!; \
	sleep 2; \
	trap 'kill $$RECV_PID 2>/dev/null || true' EXIT; \
	go run ./sender; \
	kill $$RECV_PID 2>/dev/null || true; \
	wait $$RECV_PID 2>/dev/null || true
	$(MAKE) down

clean:
	rm -f output.mcap

