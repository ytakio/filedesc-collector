.PHONY: all build run clean

BINARY := mcapwriter
OUTPUT := output.mcap
BUILD_DIR := build

all: build

build:
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && cmake .. && $(MAKE)

run: build
	@rm -f $(BUILD_DIR)/$(OUTPUT)
	cd $(BUILD_DIR) && ./$(BINARY)

clean:
	rm -rf $(BUILD_DIR) $(OUTPUT)

