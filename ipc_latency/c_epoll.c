#include <time.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/epoll.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <pthread.h>

typedef struct four_tuple {
    unsigned int src_address;
    unsigned int dst_address;
    unsigned short src_port;
    unsigned short dst_port;
} FourTuple;

typedef struct channel_data {
    int packet_type;
    int payload_len;
    int *pointer;
    FourTuple f;
    long time_;
    
} ChannelData;

#define TIMES 1000
#define MAX_EVENTS 50
#define N_THREAD 1

long latency[N_THREAD][TIMES];
const u_int16_t port = 50010;

int setnonblocking(int fd) {
    int flags = fcntl(fd, F_GETFL, 0);
    if (flags == -1) {
        perror("setnonblocking");
        return -1;
    }

    flags |= O_NONBLOCK;

    return fcntl(fd, F_SETFL, flags);
}

void server() {
    struct epoll_event ev, events[MAX_EVENTS];
    int listen_sock, conn_sock, nfds, epollfd;
    struct sockaddr_in server_address; memset(&server_address, 0, sizeof(struct sockaddr_in));
    ChannelData rxdata;
    struct timespec ts;

    // Setup server's address information
    server_address.sin_family = AF_INET;
    server_address.sin_port = htons(port);
    server_address.sin_addr.s_addr = INADDR_ANY;

    // Create listen socket
    listen_sock = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);

    // Set reuse option
    int reuse = 1;
    setsockopt(listen_sock, SOL_SOCKET, SO_REUSEADDR, (const char*)&reuse, sizeof(reuse));
    // bind and listen
    bind(listen_sock, (const struct sockaddr *) &server_address, sizeof(struct sockaddr_in));
    listen(listen_sock, 5);

    // Create epoll instance
    epollfd = epoll_create1(0);
    if (epollfd == -1) {
        perror("epoll_create1");
        exit(EXIT_FAILURE);
    }

    ev.events = EPOLLIN;
    ev.data.fd = listen_sock;
    if (epoll_ctl(epollfd, EPOLL_CTL_ADD, listen_sock, &ev) == -1) {
        perror("epoll_ctl: listen_sock");
        exit(EXIT_FAILURE);
    }

    // Start polling
    for(;;) {
        nfds = epoll_wait(epollfd, events, MAX_EVENTS, -1);
        if (nfds == -1) {
            perror("epoll_wait");
            exit(EXIT_FAILURE);
        }

        for(int x=0, n=0; x < nfds; ++x) {
            if (events[x].data.fd == listen_sock) {
                // Accept a client connnection
                conn_sock = accept(listen_sock, NULL, NULL);
                if (conn_sock == -1) {
                    perror("accept");
                    exit(EXIT_FAILURE);
                }

                // setnonblocking(conn_sock);
                // ev.events = EPOLLIN | EPOLLET;   // Edge-Trigger
                ev.events = EPOLLIN;    // Level-Trigger
                ev.data.fd = conn_sock;
                if (epoll_ctl(epollfd, EPOLL_CTL_ADD, conn_sock, &ev) == -1) {
                    perror("epoll_ctl: conn_sock");
                    exit(EXIT_FAILURE);
                }
            } else {
                // Handle client
                n = read(events[x].data.fd, &rxdata, sizeof(rxdata));
                if (n > 0) {
                    clock_gettime(CLOCK_MONOTONIC, &ts);
                    write(events[x].data.fd, &rxdata, sizeof(rxdata));
                }
            }
        }
    }
}

void *client(void *arg) {
    struct sockaddr_in server_address; memset(&server_address, 0, sizeof(struct sockaddr_in));
    ChannelData txdata;
    struct timespec start, end;
    int thread_id = *((int *)arg);

    int sock = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    server_address.sin_family = AF_INET;
    server_address.sin_port = htons(port);
    inet_pton(AF_INET, "127.0.0.1", &server_address.sin_addr);

    connect(sock, (const struct sockaddr *)&server_address, sizeof(struct sockaddr_in));

    for (int x=0; x<TIMES; ++x) {
        clock_gettime(CLOCK_MONOTONIC, &start);
        txdata.time_ = start.tv_nsec;

        write(sock, &txdata, sizeof(txdata));
        read(sock, &txdata, sizeof(txdata));
        
        clock_gettime(CLOCK_MONOTONIC, &end);
        
        latency[thread_id][x] = end.tv_nsec - start.tv_nsec;  // Round-Trip
    }
    
    close(sock);
    pthread_exit(NULL);
}

int main(int argc, char *argv[]) {
    if (*(argv[1]) == 's') {
        server();

    } else {
        int *tids = (int *) malloc(sizeof(int) * N_THREAD);
        pthread_t client_threads[N_THREAD];

        for (int x=0; x < N_THREAD; ++x) {
            tids[x] = x;

            if (pthread_create(&client_threads[x], NULL, client, (void *) &tids[x]) != 0) {
                perror("Create client thread");
                return -1;
            }
        }
        for (int x=0; x < N_THREAD; ++x) {
            pthread_join(client_threads[x], NULL);
        }
    }

    long long s = 0;
    FILE *fp = fopen("latency.csv", "w");
    for(int x=0; x < N_THREAD; ++x) {
        for (int y=0; y < TIMES; ++y) {
            s += latency[x][y];

            if (y != TIMES-1) {
                fprintf(fp, "%ld,", latency[x][y]);
            } else {
                fprintf(fp, "%ld\n", latency[x][y]);
            }
        }
    }
    fclose(fp);
    printf("Avg. %lld (ns)\n", s/(N_THREAD * TIMES));
}