.PHONY: all build run clean protoc pbgo

EXAMPLE_DIR := example/mcapwriter
BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin

# Repo-root directory (this Makefile is at go/). This example is at
# go/example/mcapwriter.
REPO_ROOT := ..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Repo-root directory where upstream .proto schemas are extracted.
PROTO_ROOT := $(REPO_ROOT)/proto

# Directory where generated Go protobuf bindings are emitted.
GEN_DIR := $(BUILD_DIR)/proto

# protoc-gen-go is declared as a Go tool in go.mod; resolve its path lazily.
PROTOC_GEN_GO = $(shell go tool -n protoc-gen-go)

# Module prefix for generated Go protobuf packages. The suffix is derived from
# each .proto file's directory under $(PROTO_ROOT) so the Makefile never
# hardcodes the upstream schema vendor name (e.g. foxglove/). Use `go list -m`
# because go/ itself contains no .go files.
GO_MODULE_PATH := $(shell go list -m -f '{{.Path}}')

all: build

build: protoc
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/ ./$(EXAMPLE_DIR)

run: protoc
	cd $(EXAMPLE_DIR) && go run .

clean:
	rm -rf $(BUILD_DIR)

# Shared cleanup target for the whole repository; removes the shared schema
# archive and extracted .proto files from the repo root.
clean-all: clean
	rm -rf $(PROTO_ROOT) $(FOXGLOVE_TARBALL)

# Extract .proto files first, then let a sub-make see them via wildcard.
protoc: | $(PROTO_ROOT)
	$(MAKE) pbgo

$(PROTO_ROOT): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_ROOT)
	@tar xzf $< -C $(PROTO_ROOT) --strip-components=3 '*/schemas/proto/*'

# Proto sources are discovered after extraction so that the Makefile can be
# re-read once the schema directory exists. No subdirectory is hardcoded;
# every .proto under $(PROTO_ROOT) is compiled.
SRCS := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)

# Map each .proto file to its Go import path. The proto import path is the
# path under $(PROTO_ROOT); the Go package suffix is the file's directory under
# $(PROTO_ROOT), so the Makefile never hardcodes the upstream schema vendor
# name (e.g. foxglove/). For a proto directly under $(PROTO_ROOT) the suffix
# becomes empty, yielding the module's build/proto package.
proto_to_go_pkg = $(GO_MODULE_PATH)/build/proto/$(patsubst ./%,%,$(patsubst %/,%,$(dir $(1))))

GO_OPTS := $(foreach src,$(SRCS),\
	--go_opt=M$(patsubst $(PROTO_ROOT)/%,%,$(src))=$(call proto_to_go_pkg,$(patsubst $(PROTO_ROOT)/%,%,$(src))))

# Sentinel that records the last successful protoc invocation. protoc produces
# many outputs from many inputs, so a single sentinel avoids rebuilding
# whenever any individual .pb.go is older than its source.
PROTOC_SENTINEL := $(GEN_DIR)/protoc-go.done

pbgo: $(PROTOC_SENTINEL)

$(PROTOC_SENTINEL): $(SRCS)
	@mkdir -p $(GEN_DIR)
	protoc \
		--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
		--proto_path=$(PROTO_ROOT) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		$(GO_OPTS) \
		$(SRCS)
	@touch $@

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

