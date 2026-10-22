.PHONY: all build run clean proto test

BINARY := mcapwriter
OUTPUT := output.mcap

# Pin the Foxglove SDK release used to source protobuf schemas.
# The proto files are downloaded at build time so we do not have to maintain
# local copies of upstream schema definitions.
FOXGLOVE_SDK_VERSION := sdk/v0.30.2
FOXGLOVE_SDK_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_SDK_VERSION).tar.gz
FOXGLOVE_SDK_TARBALL := foxglove-sdk-$(subst /,-,$(FOXGLOVE_SDK_VERSION)).tar.gz
FOXGLOVE_SDK_DIR := foxglove-sdk-sdk-v0.30.2
FOXGLOVE_PROTO_SRC := $(FOXGLOVE_SDK_DIR)/schemas/proto/foxglove
FOXGLOVE_PROTO_DIR := foxglove
GO_PACKAGE := github.com/ytakio/filedesc-collector/examples/mcapwriter/foxglove

all: build

build: proto
	go build -o $(BINARY) .

run: proto
	go run .

clean:
	rm -f $(BINARY) $(OUTPUT)
	rm -rf $(FOXGLOVE_PROTO_DIR) $(FOXGLOVE_SDK_DIR) $(FOXGLOVE_SDK_TARBALL)

# Download the pinned Foxglove SDK release, patch the proto files with the
# correct go_package option for this module, and generate Go bindings.
proto:
	@mkdir -p $(FOXGLOVE_PROTO_DIR)
	@if [ ! -f "$(FOXGLOVE_SDK_TARBALL)" ]; then \
		echo "Downloading Foxglove SDK $(FOXGLOVE_SDK_VERSION)..."; \
		curl -fsSL -o "$(FOXGLOVE_SDK_TARBALL)" "$(FOXGLOVE_SDK_URL)"; \
	fi
	@rm -rf "$(FOXGLOVE_SDK_DIR)"
	@tar -xzf "$(FOXGLOVE_SDK_TARBALL)"
	@cp "$(FOXGLOVE_PROTO_SRC)"/*.proto "$(FOXGLOVE_PROTO_DIR)/"
	@rm -rf "$(FOXGLOVE_SDK_DIR)"
	@echo "Patching go_package options..."
	@for f in "$(FOXGLOVE_PROTO_DIR)"/*.proto; do \
		if ! grep -q 'option go_package' "$$f"; then \
			awk '/^package foxglove;/ && !seen { print; print "option go_package = \"$(GO_PACKAGE)\";"; seen = 1; next } { print }' "$$f" > "$$f.tmp"; \
			mv "$$f.tmp" "$$f"; \
		fi; \
	done
	protoc --go_out=. --go_opt=paths=source_relative "$(FOXGLOVE_PROTO_DIR)"/*.proto

test:
	go test ./...

