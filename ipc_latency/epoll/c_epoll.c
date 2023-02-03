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
#include <math.h>

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

#define TIMES 5000
#define MAX_EVENTS 4096
#define N_THREAD 512

long latency[N_THREAD][TIMES];
const u_int16_t port = 50012; // select
// const u_int16_t port = 50013; // epoll

void get_monotonic_time(struct timespec* ts) {
    clock_gettime(CLOCK_MONOTONIC, ts);
}

long get_time_nano(struct timespec* ts) {
    return (long)ts->tv_sec * 1e9 + ts->tv_nsec;
}

long get_elapsed_time_nano(struct timespec* before, struct timespec* after) {
    return get_time_nano(after) - get_time_nano(before);
}

int setnonblocking(int fd) {
    int flags = fcntl(fd, F_GETFL, 0);
    if (flags == -1) {
        perror("setnonblocking");
        return -1;
    }

    flags |= O_NONBLOCK;

    return fcntl(fd, F_SETFL, flags);
}

double calculateSD(long data[]) {
    double sum = 0.0, mean, SD = 0.0;
    int i;
    for (i = 0; i < TIMES; ++i) {
        sum += (double) data[i];
    }
    mean = sum / TIMES;
    for (i = 0; i < TIMES; ++i) {
        SD += pow(data[i] - mean, 2);
    }
    return sqrt(SD / TIMES);
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
        // printf("nfds: %d\n", nfds);

        for(int x=0, n=0; x < nfds; ++x) {
            if (events[x].data.fd == listen_sock) {
                // Accept a client connnection
                conn_sock = accept(listen_sock, NULL, NULL);
                if (conn_sock == -1) {
                    perror("accept");
                    exit(EXIT_FAILURE);
                }

                setnonblocking(conn_sock);
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
                } else {
                    close(events[x].data.fd);
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

    if (connect(sock, (const struct sockaddr *)&server_address, sizeof(struct sockaddr_in)) < 0) {
        perror("Connect failed");
    }

    for (int x=0; x<TIMES; ++x) {
        clock_gettime(CLOCK_MONOTONIC, &start);
        txdata.time_ = start.tv_nsec;

        if (write(sock, &txdata, sizeof(txdata)) < 0) {
            perror("Write failed");
        }
        if (read(sock, &txdata, sizeof(txdata)) < 0) {
            perror("Read failed");
        }
        
        clock_gettime(CLOCK_MONOTONIC, &end);
        
        latency[thread_id][x] = get_elapsed_time_nano(&start, &end);  // Round-Trip
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
    // FILE *fp = fopen("latency.csv", "w");
    for(int x=0; x < N_THREAD; ++x) {
        double tmp = 0;
        for (int y=0; y < TIMES; ++y) {
            s += latency[x][y];
            tmp += latency[x][y];

            // if (y != TIMES-1) {
            //     fprintf(fp, "%ld,", latency[x][y]);
            // } else {
            //     fprintf(fp, "%ld\n", latency[x][y]);
            // }
        }
        fprintf(stderr, "%3d: SD: %.3f\tAvg: %.3f\n", x, calculateSD(latency[x]), tmp/TIMES);
    }
    // fclose(fp);
    // printf("Avg. %lld (ns)\n", s/(N_THREAD * TIMES));
    printf("%.3f\n", s/(N_THREAD * TIMES)/1000.0);
}