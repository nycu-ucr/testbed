#include <errno.h>
#include <fcntl.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include <sys/poll.h>
#include <sys/socket.h>
#include <unistd.h>
#include <time.h>
#include <sys/un.h>

#include "liburing.h"

#define MAX_CONNECTIONS     4096
#define MAX_MESSAGE_LEN     64
#define BUFFERS_COUNT       MAX_CONNECTIONS
#define ROUNDTRIP_TIMES     1000
#define USE_TCP             0
#define USE_UDS             1
#define UDS_NAME            "/tmp/uds"

void add_socket_read(struct io_uring *ring, int fd, unsigned gid, size_t size, unsigned flags);
void add_socket_write(struct io_uring *ring, int fd, __u16 bid, size_t size, unsigned flags);
void add_provide_buf(struct io_uring *ring, __u16 bid, unsigned gid);

void get_monotonic_time(struct timespec* ts) {
    clock_gettime(CLOCK_MONOTONIC, ts);
}

long get_time_nano(struct timespec* ts) {
    return (long)ts->tv_sec * 1e9 + ts->tv_nsec;
}

long get_elapsed_time_nano(struct timespec* before, struct timespec* after) {
    return get_time_nano(after) - get_time_nano(before);
}

enum {
    ACCEPT,
    READ,
    WRITE,
    PROV_BUF,
};

typedef struct conn_info {
    __u32 fd;
    __u16 type;
    __u16 bid;
} conn_info;

typedef struct channel_data {
    char buf[MAX_MESSAGE_LEN];
    struct timespec time_;
    
} ChannelData;

ChannelData bufs[BUFFERS_COUNT] = {0};
int group_id = 1337;
long latency[ROUNDTRIP_TIMES] = {0};
int message_size = sizeof(ChannelData);

int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("Please give a port number: ./io_uring_echo_server [port]\n");
        exit(0);
    }

    fprintf(stderr, "Message size: %d\n", message_size);

    // setup socket
    #if USE_TCP
        int portno = strtol(argv[1], NULL, 10);
        int client_sock = socket(AF_INET, SOCK_STREAM, 0);
        struct sockaddr_in serv_addr;

        memset(&serv_addr, 0, sizeof(serv_addr));
        serv_addr.sin_family = AF_INET;
        serv_addr.sin_port = htons(portno);
        serv_addr.sin_addr.s_addr = INADDR_ANY;

        if (connect(client_sock, (const struct sockaddr *) &serv_addr, strlen(serv_addr)) < 0) {
            perror("Connect failed");
            exit(1);
        } 
    #elif USE_UDS
        int client_sock = socket(AF_UNIX, SOCK_STREAM, 0);
        struct sockaddr_un serv_addr;

        memset(&serv_addr, 0, sizeof(struct sockaddr_un));
        serv_addr.sun_family = AF_UNIX;
        strcpy(serv_addr.sun_path, UDS_NAME);
        
        if (connect(client_sock, (const struct sockaddr *)&serv_addr, sizeof(struct sockaddr_un)) != 0) {
            perror("Connection failed");
            exit(1);
        }
    #endif

    // initialize io_uring
    struct io_uring_params params;
    struct io_uring ring;
    memset(&params, 0, sizeof(params));

    if (io_uring_queue_init_params(2048, &ring, &params) < 0) {
        perror("io_uring_init_failed...\n");
        exit(1);
    }

    // check if IORING_FEAT_FAST_POLL is supported
    if (!(params.features & IORING_FEAT_FAST_POLL)) {
        printf("IORING_FEAT_FAST_POLL not available in the kernel, quiting...\n");
        exit(0);
    }

    // check if buffer selection is supported
    struct io_uring_probe *probe;
    probe = io_uring_get_probe_ring(&ring);
    if (!probe || !io_uring_opcode_supported(probe, IORING_OP_PROVIDE_BUFFERS)) {
        printf("Buffer select not supported, skipping...\n");
        exit(0);
    }
    // free(probe);

    // register buffers for buffer selection
    struct io_uring_sqe *sqe;
    struct io_uring_cqe *cqe;

    sqe = io_uring_get_sqe(&ring);
    io_uring_prep_provide_buffers(sqe, bufs, MAX_MESSAGE_LEN, BUFFERS_COUNT, group_id, 0);

    io_uring_submit(&ring);
    io_uring_wait_cqe(&ring, &cqe);
    if (cqe->res < 0) {
        printf("cqe->res = %d\n", cqe->res);
        exit(1);
    }
    io_uring_cqe_seen(&ring, cqe);

    add_socket_write(&ring, client_sock, 4095, message_size, 0);  // bid is always 4095
    // start event loop
    int index = 0;
    for(; index != ROUNDTRIP_TIMES ;) {
        io_uring_submit_and_wait(&ring, 1);
        struct io_uring_cqe *cqe;
        unsigned head;
        unsigned count = 0;

        // go through all CQEs
        io_uring_for_each_cqe(&ring, head, cqe) {
            ++count;
            struct conn_info conn_i;
            memcpy(&conn_i, &cqe->user_data, sizeof(conn_i));

            int type = conn_i.type;
            if (cqe->res == -ENOBUFS) {
                fprintf(stdout, "bufs in automatic buffer selection empty, this should not happen...\n");
                fflush(stdout);
                exit(1);
            } else if (type == PROV_BUF) {
                if (cqe->res < 0) {
                    printf("cqe->res = %d\n", cqe->res);
                    exit(1);
                }
            } else if (type == READ) {
                int bytes_read = cqe->res;
                int bid = cqe->flags >> 16;
                if (cqe->res <= 0) {
                    // read failed, re-add the buffer
                    add_provide_buf(&ring, bid, group_id);
                    // connection closed or error
                    close(conn_i.fd);
                } else {
                    struct timespec now;

                    get_monotonic_time(&now);
                    latency[index] = get_elapsed_time_nano(&bufs[bid].time_, &now);
                    ++index;

                    // bytes have been read into bufs, now add write to socket sqe
                    add_socket_write(&ring, conn_i.fd, bid, bytes_read, 0);
                }
            } else if (type == WRITE) {
                // write has been completed, first re-add the buffer
                add_provide_buf(&ring, conn_i.bid, group_id);
                // add a new read for the existing connection
                add_socket_read(&ring, conn_i.fd, group_id, message_size, IOSQE_BUFFER_SELECT);
            }
        }

        io_uring_cq_advance(&ring, count);
    }

    long sum = 0;
    for(int x=0; x < ROUNDTRIP_TIMES; ++x) {
        sum += latency[x];
        // printf("%3d: %ld\n", x, latency[x]);
    }
    printf("Avg. latency: %.3f (us)\n", sum/ROUNDTRIP_TIMES/1000.0);
}

void add_socket_read(struct io_uring *ring, int fd, unsigned gid, size_t message_size, unsigned flags) {
    // printf("Do read\n");
    struct io_uring_sqe *sqe = io_uring_get_sqe(ring);
    io_uring_prep_recv(sqe, fd, NULL, message_size, 0);
    io_uring_sqe_set_flags(sqe, flags);
    sqe->buf_group = gid;

    conn_info conn_i = {
        .fd = fd,
        .type = READ,
    };
    memcpy(&sqe->user_data, &conn_i, sizeof(conn_i));
}

void add_socket_write(struct io_uring *ring, int fd, __u16 bid, size_t message_size, unsigned flags) {
    // printf("Do write\n");
    struct io_uring_sqe *sqe = io_uring_get_sqe(ring);

    get_monotonic_time(&bufs[bid].time_);

    io_uring_prep_send(sqe, fd, &bufs[bid], message_size, 0);
    io_uring_sqe_set_flags(sqe, flags);

    conn_info conn_i = {
        .fd = fd,
        .type = WRITE,
        .bid = bid,
    };
    memcpy(&sqe->user_data, &conn_i, sizeof(conn_i));
}

void add_provide_buf(struct io_uring *ring, __u16 bid, unsigned gid) {
    // printf("add_provide_buf: bid: %d\n", bid);
    struct io_uring_sqe *sqe = io_uring_get_sqe(ring);
    io_uring_prep_provide_buffers(sqe, &bufs[bid], MAX_MESSAGE_LEN, 1, gid, bid);

    conn_info conn_i = {
        .fd = 0,
        .type = PROV_BUF,
    };
    memcpy(&sqe->user_data, &conn_i, sizeof(conn_i));
}
