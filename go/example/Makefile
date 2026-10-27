.PHONY: all example test clean clean-all protoc pbgo

# This Makefile is at go/example/. The repository root is two levels up.
REPO_ROOT := ../..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Directory where the upstream SDK source is extracted.
FOXGLOVE_SDK_DIR := $(REPO_ROOT)/foxglove-sdk

# Directory containing upstream .proto schemas.
PROTO_ROOT := $(FOXGLOVE_SDK_DIR)/schemas/proto

# Directory where generated Go protobuf bindings are emitted.
GEN_DIR := proto

# protoc-gen-go is declared as a Go tool in go.mod; resolve its path lazily.
PROTOC_GEN_GO = $(shell go tool -n protoc-gen-go)

# Module prefix for generated Go protobuf packages. The suffix is derived from
# each .proto file's directory under $(PROTO_ROOT) so the Makefile never
# hardcodes the upstream schema vendor name (e.g. foxglove/). Use `go list -m`
# because go/ itself contains no .go files.
GO_MODULE_PATH := $(shell go list -m -f '{{.Path}}')

all: example

example: protoc
	go run ./mcapwriter

clean:
	rm -rf $(GEN_DIR)

# Shared cleanup target for the whole repository; removes the shared SDK
# source tree and archive from the repo root.
clean-all: clean
	rm -rf $(FOXGLOVE_SDK_DIR) $(FOXGLOVE_TARBALL)

# Extract the SDK source tree first, then let a sub-make discover the proto
# files via find. Using --strip-components=1 instead of --wildcards works on
# both GNU tar and BSD tar.
protoc: | $(FOXGLOVE_SDK_DIR)
	$(MAKE) pbgo

$(FOXGLOVE_SDK_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(FOXGLOVE_SDK_DIR)
	@tar xzf $< -C $(FOXGLOVE_SDK_DIR) --strip-components=1

# Proto sources are discovered after extraction so that the Makefile can be
# re-read once the schema directory exists. No subdirectory is hardcoded;
# every .proto under $(PROTO_ROOT) is compiled.
PROTO_SRCS := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)

# Build --go_opt flags by joining per-proto pieces. Splitting the work into
# small named lists is much easier to follow than a single nested expression.
PROTO_RELPATHS := $(patsubst $(PROTO_ROOT)/%,%,$(PROTO_SRCS))
PROTO_RELDIRS  := $(patsubst %/,%,$(dir $(PROTO_RELPATHS)))
GO_PKG_SUFFIX  := $(patsubst ./%,%,$(PROTO_RELDIRS))
GO_PKG_PATHS   := $(addprefix $(GO_MODULE_PATH)/example/proto/,$(GO_PKG_SUFFIX))

PROTOC_GO_OPTS := $(join \
	$(addprefix --go_opt=M,$(PROTO_RELPATHS)),\
	$(addprefix =,$(GO_PKG_PATHS)))

# Generated .pb.go files, one per proto source. All are produced by a single
# protoc invocation because the --go_opt=M mappings depend on the full set of
# inputs.
GO_SRCS := $(patsubst $(PROTO_ROOT)/%.proto,$(GEN_DIR)/%.pb.go,$(PROTO_SRCS))

pbgo: $(GO_SRCS)

$(GO_SRCS): $(PROTO_SRCS)
	@mkdir -p $(GEN_DIR)
	protoc \
		--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
		--proto_path=$(PROTO_ROOT) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		$(PROTOC_GO_OPTS) \
		$(PROTO_SRCS)

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

