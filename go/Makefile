.PHONY: all example test clean clean-all

all: test

test:
	go test ./filedesc

example:
	$(MAKE) -C example example

clean:
	$(MAKE) -C example clean

clean-all: clean
	$(MAKE) -C example clean-all

