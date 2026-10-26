.PHONY: all build run clean protoc pbgen

EXAMPLE_DIR := example/mcapwriter
BINARY := mcapwriter
BUILD_DIR := build/example/mcapwriter
OUTPUT := $(BUILD_DIR)/output.mcap

# Repo-root directory (this Makefile is at cpp/). This example is at
# cpp/example/mcapwriter.
REPO_ROOT := ..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Repo-root directory where upstream .proto schemas are extracted.
PROTO_ROOT := $(REPO_ROOT)/proto

# Directory where generated C++ protobuf bindings are emitted.
GEN_DIR := build/proto

all: build

build: protoc
	cmake -B $(BUILD_DIR) -S $(EXAMPLE_DIR)
	cmake --build $(BUILD_DIR)

run: build
	@rm -f $(OUTPUT)
	cd $(BUILD_DIR) && ./$(BINARY)

clean:
	rm -rf build

# Shared cleanup target for the whole repository; removes the shared schema
# archive and extracted .proto files from the repo root.
clean-all: clean
	rm -rf $(PROTO_ROOT) $(FOXGLOVE_TARBALL)

# Extract upstream .proto files into the repo-root proto/ directory, then
# generate C++ bindings into cpp/build/proto/. Because a single protoc
# invocation produces every output, use a stamp file to avoid re-running
# whenever an individual generated file is older than its source. The stamp
# depends on the tarball (a stable file) and uses $(PROTO_ROOT) as an
# order-only prerequisite so the directory is created without its mtime
# triggering rebuilds.
protoc: | $(PROTO_ROOT)
	$(MAKE) pbgen

$(PROTO_ROOT): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_ROOT)
	@tar xzf $< -C $(PROTO_ROOT) --strip-components=3 '*/schemas/proto/*'

# Proto sources are discovered after extraction so that the Makefile can be
# re-read once the schema directory exists. No subdirectory is hardcoded;
# every .proto under $(PROTO_ROOT) is compiled.
SRCS := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)

PROTOC_SENTINEL := $(GEN_DIR)/protoc-cpp.done

pbgen: $(PROTOC_SENTINEL)

$(PROTOC_SENTINEL): $(SRCS)
	@mkdir -p $(GEN_DIR)
	protoc --proto_path=$(PROTO_ROOT) --cpp_out=$(GEN_DIR) $(SRCS)
	@touch $@

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

