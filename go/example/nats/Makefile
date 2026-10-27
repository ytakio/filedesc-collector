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
	cd $(GO_MODULE_ROOT) && go run ./example/nats/sender

receiver: protoc
	cd $(GO_MODULE_ROOT) && go run ./example/nats/receiver -output=example/nats/output.mcap

# Run the full example: start NATS, run the receiver in the background, run the
# sender, wait for the receiver to finish, then tear NATS down.
example: protoc up
	cd $(GO_MODULE_ROOT) && go run ./example/nats/receiver -output=example/nats/output.mcap &
	RECV_PID=$$!; \
	sleep 2; \
	trap 'kill $$RECV_PID 2>/dev/null || true' EXIT; \
	cd $(GO_MODULE_ROOT) && go run ./example/nats/sender; \
	wait $$RECV_PID
	$(MAKE) down

clean:
	rm -f output.mcap

