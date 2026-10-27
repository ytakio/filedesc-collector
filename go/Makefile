.PHONY: all example test clean clean-all nats-example

all: test

test:
	go test ./filedesc

example:
	$(MAKE) -C example example

nats-example:
	$(MAKE) -C example nats-example

clean:
	$(MAKE) -C example clean

clean-all: clean
	$(MAKE) -C example clean-all

