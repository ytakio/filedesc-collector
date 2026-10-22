.PHONY: all build run clean proto test

BINARY := mcapwriter
OUTPUT := output.mcap

# Foxglove SDK release tag from which to download protobuf schemas.
FOXGLOVE_TAG := ros-v3.6.3
FOXGLOVE_TARBALL := foxglove-sdk-$(FOXGLOVE_TAG).tar.gz
FOXGLOVE_DOWNLOAD_URL := https://github.com/foxglove/foxglove-sdk/archive/refs/tags/$(FOXGLOVE_TAG).tar.gz

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

# Download the Foxglove SDK release tarball, extract the protobuf schemas,
# patch their go_package option, and generate Go bindings alongside the .proto files.
proto: $(PROTO_OUT_DIR)/.proto-extracted
	@protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_DIR) \
		--go_opt=paths=source_relative \
		$(PROTO_OUT_DIR)/*.proto

$(PROTO_OUT_DIR)/.proto-extracted: $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_OUT_DIR)
	@tar xzf $< -C $(PROTO_OUT_DIR) --strip-components=4 'foxglove-sdk-$(FOXGLOVE_TAG)/schemas/proto/foxglove/*.proto'
	@echo "Patching go_package options..."
	@for f in $(PROTO_OUT_DIR)/*.proto; do \
		if ! grep -q 'option go_package' "$$f"; then \
			echo 'option go_package = "$(GO_PACKAGE)";' >> "$$f"; \
		else \
			sed -i.bak 's|option go_package = ".*";|option go_package = "$(GO_PACKAGE)";|' "$$f"; \
			rm -f "$$f.bak"; \
		fi; \
	done
	@touch $@

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_DOWNLOAD_URL)

test:
	go test ./...

