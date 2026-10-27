.PHONY: all example test clean clean-all

all: test

test:
	uv run python -m pytest

example:
	$(MAKE) -C example example

clean:
	$(MAKE) -C example clean

# Shared cleanup target for the whole repository; removes the shared SDK
# source tree and archive from the repo root.
clean-all: clean
	$(MAKE) -C example clean-all

