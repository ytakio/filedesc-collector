# Top-level orchestration for the filedesc-collector polyglot repository.
# Each language lives in its own subdirectory with its own build system.

.PHONY: all test example clean clean-all

all: test

test:
	$(MAKE) -C go test
	$(MAKE) -C python test
	$(MAKE) -C cpp test

example:
	$(MAKE) -C go example
	$(MAKE) -C python example
	$(MAKE) -C cpp example

clean:
	$(MAKE) -C go clean
	$(MAKE) -C python clean
	$(MAKE) -C cpp clean

clean-all:
	$(MAKE) -C go clean-all
	$(MAKE) -C python clean-all
	$(MAKE) -C cpp clean-all

