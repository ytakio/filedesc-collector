.PHONY: all example test clean clean-all protoc pbpy

EXAMPLE_DIR := example/mcapwriter

# Repo-root directory (this Makefile is at python/). This example is at
# python/example/mcapwriter.
REPO_ROOT := ..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Directory where the upstream SDK source is extracted.
FOXGLOVE_SDK_DIR := $(REPO_ROOT)/foxglove-sdk

# Directory containing upstream .proto schemas.
PROTO_ROOT := $(FOXGLOVE_SDK_DIR)/schemas/proto

# Directory under python/ where generated Python protobuf bindings are emitted.
GEN_DIR := proto

all: test

test:
	uv run python -m pytest

example: protoc
	uv run python $(EXAMPLE_DIR)/main.py

clean:
	rm -rf $(GEN_DIR)

# Shared cleanup target for the whole repository; removes the shared SDK
# source tree and archive from the repo root.
clean-all: clean
	rm -rf $(FOXGLOVE_SDK_DIR) $(FOXGLOVE_TARBALL)

# Extract the SDK source tree first, then let a sub-make discover the proto
# files via find. Using --strip-components=1 instead of a member pattern
# works on both GNU tar and BSD tar.
protoc: | $(FOXGLOVE_SDK_DIR)
	$(MAKE) pbpy

$(FOXGLOVE_SDK_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(FOXGLOVE_SDK_DIR)
	@tar xzf $< -C $(FOXGLOVE_SDK_DIR) --strip-components=1

# Proto sources are discovered after extraction so that the Makefile can be
# re-read once the schema directory exists. No subdirectory is hardcoded;
# every .proto under $(PROTO_ROOT) is compiled.
PROTO_SRCS := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)

# Generated Python protobuf bindings, one per proto source.
PY_SRCS := $(patsubst $(PROTO_ROOT)/%.proto,$(GEN_DIR)/%_pb2.py,$(PROTO_SRCS))

pbpy: $(PY_SRCS)

$(GEN_DIR)/%_pb2.py: $(PROTO_ROOT)/%.proto
	@mkdir -p $(dir $@)
	protoc \
		--proto_path=$(PROTO_ROOT) \
		--python_out=$(GEN_DIR) \
		$<
	@sed -i.bak -e 's/from foxglove import \([A-Za-z_][A-Za-z0-9_]*\) as/from . import \1 as/g' $@
	@rm -f $@.bak

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

