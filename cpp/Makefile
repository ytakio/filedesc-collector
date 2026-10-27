.PHONY: all example test clean clean-all

all: test

test:
	cmake -B build -S . -DBUILD_EXAMPLES=OFF
	cmake --build build
	ctest --test-dir build --output-on-failure

example:
	$(MAKE) -C example example

clean:
	$(MAKE) -C example clean

clean-all: clean
	$(MAKE) -C example clean-all

