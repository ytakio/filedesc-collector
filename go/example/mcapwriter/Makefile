.PHONY: all build run clean protoc

OUT_DIR := bin
OUTPUT := output.mcap

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

PROTO_DIR := proto
GO_MODULE := $(shell go list -f '{{.Module.Path}}')

# protoc-gen-go is declared as a Go tool in go.mod; resolve its path lazily.
PROTOC_GEN_GO = $(shell go tool -n protoc-gen-go)

all: build

build: protoc
	go build -o $(OUT_DIR)/ .

run: protoc
	go run .

clean:
	rm -rf $(OUT_DIR) $(OUTPUT)
	rm -rf $(PROTO_DIR)

# Extract .proto files first, then let a sub-make see them via wildcard.
protoc: | $(PROTO_DIR)
	$(MAKE) pbgo

$(PROTO_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_DIR)
	@tar xzf $< -C $(PROTO_DIR) --strip-components=3 '*/schemas/proto/*'

# C-style source/object lists, evaluated after extraction.
SRCS := $(shell find $(PROTO_DIR) -name '*.proto' 2>/dev/null)
OBJS := $(SRCS:.proto=.pb.go)

# Map each .proto file to its Go import path. The proto import path is the
# path under $(PROTO_DIR); the Go package is $(GO_MODULE)/<schema-dir>.
GO_OPTS := $(foreach src,$(SRCS),\
	--go_opt=M$(patsubst $(PROTO_DIR)/%/,%,$(src)=$(GO_MODULE)/$(dir $(src))))

.PHONY: pbgo
pbgo: $(OBJS)

# protoc generates all .pb.go files in one invocation.
$(OBJS): $(SRCS)
	protoc \
		--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_DIR) \
		--go_opt=paths=source_relative \
		$(GO_OPTS) \
		$(SRCS)

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

