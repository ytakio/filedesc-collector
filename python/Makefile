.PHONY: all build run clean clean-all protoc

EXAMPLE_DIR := example/mcapwriter
OUTPUT := $(EXAMPLE_DIR)/output.mcap

# Repo-root directory (this Makefile is at python/). This example is at
# python/example/mcapwriter.
REPO_ROOT := ..

# Shared proto archive and directory used by the C++ and Go examples.
# Python uses the foxglove-sdk package and does not generate bindings, but
# the clean-all target removes these shared artifacts for consistency.
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
PROTO_ROOT := $(REPO_ROOT)/proto

all: run

build: protoc
	@echo "Python example uses foxglove-sdk packages; no build step required."

run:
	cd $(EXAMPLE_DIR) && python main.py

clean:
	rm -f $(OUTPUT)

# Shared cleanup target for the whole repository; removes the shared schema
# archive and extracted .proto files from the repo root.
clean-all: clean
	rm -rf $(PROTO_ROOT) $(FOXGLOVE_TARBALL)

# No-op: Python does not need generated protobuf bindings from .proto files.
protoc:
	@true

