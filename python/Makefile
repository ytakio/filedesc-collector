.PHONY: all example test clean clean-all protoc pbpy

EXAMPLE_DIR := example/mcapwriter

# Repo-root directory (this Makefile is at python/). This example is at
# python/example/mcapwriter.
REPO_ROOT := ..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Repo-root directory where upstream .proto schemas are extracted.
PROTO_ROOT := $(REPO_ROOT)/proto

# Directory under python/ where generated Python protobuf bindings are emitted.
GEN_DIR := proto

all: test

test:
	uv run pytest

example: protoc
	uv run python $(EXAMPLE_DIR)/main.py

clean:
	rm -rf $(GEN_DIR)

# Shared cleanup target for the whole repository; removes the shared schema
# archive and extracted .proto files from the repo root.
clean-all: clean
	rm -rf $(PROTO_ROOT) $(FOXGLOVE_TARBALL)

# Extract .proto files first, then let a sub-make see them via wildcard.
protoc: | $(PROTO_ROOT)
	$(MAKE) pbpy

$(PROTO_ROOT): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_ROOT)
	@tar xzf $< -C $(PROTO_ROOT) --wildcards --strip-components=3 '*/schemas/proto/*'

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

