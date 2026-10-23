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

# Directory under go/ where generated Go protobuf bindings are emitted.
GEN_DIR := proto

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
	rm -rf $(BUILD_DIR) $(GEN_DIR)

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
PROTO_SRCS := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)

# Build --go_opt flags by joining per-proto pieces. Splitting the work into
# small named lists is much easier to follow than a single nested expression.
PROTO_RELPATHS := $(patsubst $(PROTO_ROOT)/%,%,$(PROTO_SRCS))
PROTO_RELDIRS  := $(patsubst %/,%,$(dir $(PROTO_RELPATHS)))
GO_PKG_SUFFIX  := $(patsubst ./%,%,$(PROTO_RELDIRS))
GO_PKG_PATHS   := $(addprefix $(GO_MODULE_PATH)/proto/,$(GO_PKG_SUFFIX))

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

