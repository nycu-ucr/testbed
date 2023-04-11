CCFLAGS ?= -Wall -O2 -D_GNU_SOURCE -luring
all_targets = io_uring_echo_server io_uring_client

.PHONY: liburing io_uring_echo_server io_uring_client

all: $(all_targets)

clean:
	rm -f $(all_targets)

liburing:
	+$(MAKE) -C ./liburing

io_uring_echo_server:
	$(CC) io_uring_echo_server.c -o ./io_uring_echo_server  ${CCFLAGS}

io_uring_client:
	$(CC) io_uring_client.c -o ./io_uring_client  ${CCFLAGS}