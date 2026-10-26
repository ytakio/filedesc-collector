.PHONY: all example test clean clean-all protoc pbcpp

EXAMPLE_DIR := example/mcapwriter

# Repo-root directory (this Makefile is at cpp/). This example is at
# cpp/example/mcapwriter.
REPO_ROOT := ..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Directory where the upstream SDK source is extracted.
FOXGLOVE_SDK_DIR := $(REPO_ROOT)/foxglove-sdk

# Directory containing upstream .proto schemas.
PROTO_ROOT := $(FOXGLOVE_SDK_DIR)/schemas/proto

# Directory under cpp/ where generated C++ protobuf bindings are emitted.
GEN_DIR := proto

# Build directory and binary name for the C++ example.
BUILD_DIR := build/example/mcapwriter
EXAMPLE_BINARY := mcapwriter
EXAMPLE_OUTPUT := $(BUILD_DIR)/output.mcap

all: test

test:
	cmake -B build -S . -DBUILD_EXAMPLES=OFF
	cmake --build build
	ctest --test-dir build --output-on-failure

example: protoc
	cmake -B $(BUILD_DIR) -S $(EXAMPLE_DIR)
	cmake --build $(BUILD_DIR)
	@rm -f $(EXAMPLE_OUTPUT)
	./$(BUILD_DIR)/$(EXAMPLE_BINARY)

clean:
	rm -rf $(GEN_DIR) build

# Shared cleanup target for the whole repository; removes the shared SDK
# source tree and archive from the repo root.
clean-all: clean
	rm -rf $(FOXGLOVE_SDK_DIR) $(FOXGLOVE_TARBALL)

# Extract the SDK source tree first, then let a sub-make discover the proto
# files via find. Using --strip-components=1 instead of --wildcards works on
# both GNU tar and BSD tar.
protoc: | $(FOXGLOVE_SDK_DIR)
	$(MAKE) pbcpp

$(FOXGLOVE_SDK_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(FOXGLOVE_SDK_DIR)
	@tar xzf $< -C $(FOXGLOVE_SDK_DIR) --strip-components=1

# Proto sources are discovered after extraction so that the Makefile can be
# re-read once the schema directory exists. No subdirectory is hardcoded;
# every .proto under $(PROTO_ROOT) is compiled.
PROTO_SRCS := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)

# Generated C++ protobuf sources, one per proto source. All are produced by a
# single protoc invocation.
CPP_SRCS := $(patsubst $(PROTO_ROOT)/%.proto,$(GEN_DIR)/%.pb.cc,$(PROTO_SRCS))

pbcpp: $(CPP_SRCS)

$(CPP_SRCS): $(PROTO_SRCS)
	@mkdir -p $(GEN_DIR)
	protoc --proto_path=$(PROTO_ROOT) --cpp_out=$(GEN_DIR) $(PROTO_SRCS)

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

