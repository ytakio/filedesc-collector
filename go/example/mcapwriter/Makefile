.PHONY: all build run clean protoc

BINARY := mcapwriter
OUTPUT := output.mcap

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

PROTO_DIR := proto
GO_MODULE := github.com/ytakio/filedesc-collector/go/example/mcapwriter

# protoc-gen-go is declared as a Go tool in go.mod; resolve its path lazily.
PROTOC_GEN_GO = $(shell go tool -n protoc-gen-go)

all: build

build: protoc
	go build -o $(BINARY) .

run: protoc
	go run .

clean:
	rm -f $(BINARY) $(OUTPUT)
	rm -rf $(PROTO_DIR)

# Extract .proto files first, then let a sub-make see them via wildcard.
protoc: | $(PROTO_DIR)
	$(MAKE) pbgo

$(PROTO_DIR): $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_DIR)
	@tar xzf $< -C $(PROTO_DIR) --strip-components=3 '*/schemas/proto/*/*.proto'

# C-style source/object lists, evaluated after extraction.
SRCS := $(wildcard $(PROTO_DIR)/*/*.proto)
OBJS := $(SRCS:.proto=.pb.go)

# Discover schema directories (e.g. proto/foxglove) and build per-directory
# --go_opt=M mappings so each schema package imports from $(GO_MODULE)/proto/<dir>.
SCHEMA_DIRS := $(sort $(patsubst %/,%,$(dir $(SRCS))))
SCHEMA_NAMES := $(notdir $(SCHEMA_DIRS))
GO_OPTS := $(foreach name,$(SCHEMA_NAMES),\
	$(foreach src,$(wildcard $(PROTO_DIR)/$(name)/*.proto),\
		--go_opt=M$(name)/$(notdir $(src))=$(GO_MODULE)/proto/$(name)))

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

