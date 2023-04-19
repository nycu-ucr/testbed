CC = gcc
CCFLAGS ?= -Wall -O2 -D_GNU_SOURCE -luring
all_targets = io_uring_echo_server io_uring_client

.PHONY: liburing io_uring_echo_server io_uring_client clean

all: $(all_targets)

clean:
	rm -f $(all_targets)
	-rm -rf ./liburing
	-rm -f /tmp/uds

liburing:
	-@git clone https://github.com/axboe/liburing.git
	+$(MAKE) -C ./liburing
	# If failed
	# cd ./liburing; sudo make install

io_uring_echo_server:
	$(CC) io_uring_echo_server.c -o ./io_uring_echo_server  ${CCFLAGS}

io_uring_client:
	$(CC) io_uring_client.c -o ./io_uring_client  ${CCFLAGS}