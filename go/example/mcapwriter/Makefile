.PHONY: all build run clean proto

BINARY := mcapwriter
OUTPUT := output.mcap

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := sdk/v0.30.2
FOXGLOVE_TARBALL := foxglove-sdk.tar.gz
FOXGLOVE_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

# Directory where downloaded .proto files and generated .pb.go files coexist.
PROTO_DIR := proto
PROTO_OUT_DIR := $(PROTO_DIR)/foxglove
GO_PACKAGE := github.com/ytakio/filedesc-collector/go/example/mcapwriter/proto/foxglove

all: build

build: proto
	go build -o $(BINARY) .

run: proto
	go run .

clean:
	rm -f $(BINARY) $(OUTPUT)
	rm -rf $(PROTO_DIR) $(FOXGLOVE_TARBALL)

# Download, extract, and compile all Foxglove protobuf schemas in one shot.
proto: $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_OUT_DIR)
	@rm -rf $(PROTO_OUT_DIR).tmp
	@mkdir -p $(PROTO_OUT_DIR).tmp
	@tar xzf $< -C $(PROTO_OUT_DIR).tmp
	@cp $(PROTO_OUT_DIR).tmp/*/schemas/proto/foxglove/*.proto $(PROTO_OUT_DIR)/
	@rm -rf $(PROTO_OUT_DIR).tmp
	@cd $(PROTO_DIR) && \
	proto_files=$$(ls foxglove/*.proto) && \
	go_opts=$$(for f in $$proto_files; do echo "--go_opt=M$$f=$(GO_PACKAGE)"; done) && \
	protoc \
		--go_prefix="go tool" \
		--proto_path=. \
		--go_out=. \
		--go_opt=paths=source_relative \
		$$go_opts \
		$$proto_files

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

