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

# Generated rules that define the real .pb.go targets.  This include is rebuilt
# whenever the tarball changes, which also re-extracts the .proto files.  GNU
# Make restarts after rebuilding an included makefile.
-include $(PROTO_OUT_DIR)/.proto-rules.mk

all: build

build: proto
	go build -o $(BINARY) .

run: proto
	go run .

clean:
	rm -f $(BINARY) $(OUTPUT)
	rm -rf $(PROTO_DIR) $(FOXGLOVE_TARBALL)

# Extract the tarball and emit a makefile fragment describing the actual
# proto/foxglove/*.proto -> proto/foxglove/*.pb.go targets.
$(PROTO_OUT_DIR)/.proto-rules.mk: $(FOXGLOVE_TARBALL)
	@mkdir -p $(PROTO_OUT_DIR)
	@rm -rf $(PROTO_OUT_DIR).tmp
	@mkdir -p $(PROTO_OUT_DIR).tmp
	@tar xzf $< -C $(PROTO_OUT_DIR).tmp
	@cp $(PROTO_OUT_DIR).tmp/*/schemas/proto/foxglove/*.proto $(PROTO_OUT_DIR)/
	@rm -rf $(PROTO_OUT_DIR).tmp
	@printf '# Auto-generated proto build rules. Do not edit.\n' > $@
	@printf 'PROTO_FILES := $$(wildcard $(PROTO_OUT_DIR)/*.proto)\n' >> $@
	@printf 'PB_GO_FILES := $$(patsubst %%.proto,%%.pb.go,$$(PROTO_FILES))\n' >> $@
	@printf '\n' >> $@
	@printf 'proto: $$(PB_GO_FILES)\n' >> $@
	@printf '\n' >> $@
	@printf '$$(PB_GO_FILES): $$(PROTO_OUT_DIR)/.generated\n' >> $@
	@printf '\t@touch $$@\n' >> $@
	@printf '\n' >> $@
	@printf '$$(PROTO_OUT_DIR)/.generated: $$(PROTO_FILES)\n' >> $@
	@printf '\t@cd $(PROTO_DIR) && \\\n' >> $@
	@printf '\tprotoc \\\n' >> $@
	@printf '\t\t--go_prefix="go tool" \\\n' >> $@
	@printf '\t\t--proto_path=. \\\n' >> $@
	@printf '\t\t--go_out=. \\\n' >> $@
	@printf '\t\t--go_opt=paths=source_relative \\\n' >> $@
	@printf '\t\t$$(foreach proto,$$(notdir $$(PROTO_FILES)),--go_opt=Mfoxglove/$$(proto)=$(GO_PACKAGE)) \\\n' >> $@
	@printf '\t\t$$(patsubst $(PROTO_DIR)/%%,%%,$$(PROTO_FILES))\n' >> $@
	@printf '\t@touch $$@\n' >> $@

$(FOXGLOVE_TARBALL):
	curl -sL -o $@ $(FOXGLOVE_URL)

