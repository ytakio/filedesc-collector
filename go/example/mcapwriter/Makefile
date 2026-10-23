.PHONY: all build run clean protoc pbgo

OUT_DIR := bin
OUTPUT := output.mcap

# Repo-root directory (this example is at go/example/mcapwriter).
REPO_ROOT := ../../..

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := $(REPO_ROOT)/foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Repo-root directory where upstream .proto schemas are extracted.
PROTO_ROOT := $(REPO_ROOT)/proto
PROTO_DIR := $(PROTO_ROOT)/foxglove

# Directory where generated Go protobuf bindings are emitted.
GEN_DIR := ../../proto

# protoc-gen-go is declared as a Go tool in go.mod; resolve its path lazily.
PROTOC_GEN_GO = $(shell go tool -n protoc-gen-go)

# Go import path for the generated foxglove protobuf package.
GO_PROTO_PACKAGE := github.com/ytakio/filedesc-collector/go/proto/foxglove

all: build

build: protoc
	go build -o $(OUT_DIR)/ .

run: protoc
	go run .

clean:
	rm -rf $(OUT_DIR) $(OUTPUT)
	rm -rf $(GEN_DIR)

# Shared cleanup target for the whole repository; removes the shared schema
# archive and extracted .proto files from the repo root.
clean-all: clean
	rm -rf $(PROTO_ROOT) $(FOXGLOVE_TARBALL)

# Extract .proto files first, then let a sub-make see them via wildcard.
protoc: | $(PROTO_DIR)
	$(MAKE) pbgo

$(PROTO_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_ROOT)
	@tar xzf $< -C $(PROTO_ROOT) --strip-components=3 '*/schemas/proto/*'

# C-style source/object lists, evaluated after extraction.
SRCS := $(shell find $(PROTO_DIR) -name '*.proto' 2>/dev/null)
OBJS := $(SRCS:$(PROTO_DIR)/%.proto=$(GEN_DIR)/%.pb.go)

# Map each .proto file to its Go import path. The proto import path is the
# path under $(PROTO_ROOT); all generated foxglove bindings live in the same
# Go package.
GO_OPTS := $(foreach src,$(SRCS),\
	--go_opt=M$(patsubst $(PROTO_ROOT)/%,%,$(src))=$(GO_PROTO_PACKAGE))

pbgo: $(OBJS)

# protoc generates all .pb.go files in one invocation. Because a single
# invocation produces every output, we use a stamp file to avoid re-running
# whenever any individual .pb.go is older than its source .proto.
$(GEN_DIR)/.stamp: $(SRCS)
	@mkdir -p $(GEN_DIR)
	protoc \
		--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
		--proto_path=$(PROTO_ROOT) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		$(GO_OPTS) \
		$(SRCS)
	@touch $@

$(OBJS): $(GEN_DIR)/.stamp ;

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

