.PHONY: all build run clean protoc

EXAMPLE_DIR := example/mcapwriter
BINARY := mcapwriter
OUTPUT := output.mcap
BUILD_DIR := $(EXAMPLE_DIR)/build

# Repo-root directory (this Makefile is at cpp/). This example is at
# cpp/example/mcapwriter.
REPO_ROOT := ..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Repo-root directory where upstream .proto schemas are extracted.
PROTO_ROOT := $(REPO_ROOT)/proto
PROTO_DIR := $(PROTO_ROOT)/foxglove

# Directory where generated C++ protobuf bindings are emitted.
GEN_DIR := proto

all: build

build: protoc
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && cmake .. && $(MAKE)

run: build
	@rm -f $(BUILD_DIR)/$(OUTPUT)
	cd $(BUILD_DIR) && ./$(BINARY)

clean:
	rm -rf $(BUILD_DIR) $(OUTPUT)
	rm -rf $(GEN_DIR)

# Shared cleanup target for the whole repository; removes the shared schema
# archive and extracted .proto files from the repo root.
clean-all: clean
	rm -rf $(PROTO_ROOT) $(FOXGLOVE_TARBALL)

# Extract upstream .proto files into the repo-root proto/ directory, then
# generate C++ bindings into cpp/proto/foxglove/. Because a single protoc
# invocation produces every output, use a stamp file to avoid re-running
# whenever an individual generated file is older than its source. The stamp
# depends on the tarball (a stable file) and uses $(PROTO_DIR) as an order-only
# prerequisite so the directory is created without its mtime triggering
# rebuilds.
protoc: $(GEN_DIR)/.stamp

$(GEN_DIR)/.stamp: $(FOXGLOVE_TARBALL) | $(PROTO_DIR)
	@mkdir -p $(GEN_DIR)
	protoc --proto_path=$(PROTO_ROOT) --cpp_out=$(GEN_DIR) $(PROTO_DIR)/*.proto
	@touch $@

$(PROTO_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_ROOT)
	@tar xzf $< -C $(PROTO_ROOT) --strip-components=3 '*/schemas/proto/*'

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

