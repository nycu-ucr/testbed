#include <time.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <sys/un.h>

#define TIMES 1000

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

double *latency;
const u_int16_t port = 50010;
int n_threads;

void get_monotonic_time(struct timespec* ts) {
    clock_gettime(CLOCK_MONOTONIC, ts);
}

long get_time_nano(struct timespec* ts) {
    return (long)ts->tv_sec * 1e9 + ts->tv_nsec;
}

long get_elapsed_time_nano(struct timespec* before, struct timespec* after) {
    return get_time_nano(after) - get_time_nano(before);
}

char *make_msg(int size) {
    char *ptr = (char *) malloc(sizeof(char) * size);
    memset(ptr, size, 'x');

    return ptr;
}

void *handle_client(void *arg) {
    int id = *((int *) arg);
    int res;
    ChannelData rxdata;

    struct sockaddr_un server_address; memset(&server_address, 0, sizeof(struct sockaddr_un));
    int listen_sock = socket(AF_UNIX, SOCK_STREAM, 0), client_sock;

    server_address.sun_family = AF_UNIX;
    char path[18];
    sprintf(path, "./unix_domain_%03d", id);
    strncpy(server_address.sun_path, path, sizeof(path));
    

    if (bind(listen_sock, (const struct sockaddr *) &server_address, sizeof(struct sockaddr_un)) != 0) {
        perror("Bind");
        exit(1);
    }
    if (listen(listen_sock, 5) != 0) {
        perror("Listen");
        exit(1);
    }
    client_sock = accept(listen_sock, NULL, NULL);

    for (int x=0; x<TIMES; ++x) {
        // read(client_sock, &rxdata, sizeof(rxdata));
        recv(client_sock, &rxdata, sizeof(rxdata), 0);
        // write(client_sock, &rxdata, sizeof(rxdata));
        send(client_sock, &rxdata, sizeof(rxdata), 0);
    }
    close(client_sock);

    pthread_exit(NULL);
}

void server() {
    pthread_t *threads = (pthread_t *) malloc(sizeof(pthread_t) * n_threads);
    int *ids = (int *) malloc(sizeof(int) * n_threads);
    for (int x=0; x < n_threads; ++x) {
        ids[x] = x;
        pthread_create(&threads[x], NULL, handle_client, &ids[x]);
    }

    for (int x=0; x < n_threads; ++x) {
        pthread_join(threads[x], NULL);
    }

    free(threads);
    free(ids);
}

void *do_client(void *arg) {
    ChannelData txdata;
    struct timespec start, end;
    int index = *((int *) arg);

    int sock = socket(AF_UNIX, SOCK_STREAM, 0);
    if (sock == -1) {
        perror("Create socket");
        exit(1);
    }
    struct sockaddr_un server_address; memset(&server_address, 0, sizeof(struct sockaddr_un));
    server_address.sun_family = AF_UNIX;
    char path[18];
    sprintf(path, "./unix_domain_%03d", index);
    strncpy(server_address.sun_path, path, sizeof(path));
    
    if (connect(sock, (const struct sockaddr *)&server_address, sizeof(struct sockaddr_un)) != 0) {
        perror("Connection failed");
        exit(1);
    }

    double sum;
    latency[index] = 0.0;
    
    for (int x=0; x<TIMES; ++x) {
        get_monotonic_time(&start);
        txdata.time_ = start.tv_nsec;

        // write(sock, &txdata, sizeof(txdata));
        send(sock, &txdata, sizeof(txdata), 0);
        // read(sock, &txdata, sizeof(txdata));
        recv(sock, &txdata, sizeof(txdata), 0);
        
        get_monotonic_time(&end);
        
        sum += get_elapsed_time_nano(&start, &end) / 1000.0;  // Round-Trip
    }

    latency[index] = sum / TIMES;
    
    close(sock);
}

void client() {
    pthread_t *threads = (pthread_t *) malloc(sizeof(pthread_t) * n_threads);
    int *ids = (int *) malloc(sizeof(int) * n_threads);

    for (int x=0; x < n_threads; ++x) {
        ids[x] = x;
        pthread_create(&threads[x], NULL, do_client, &ids[x]);
        // usleep(1000);
    }

    for (int x=0; x < n_threads; ++x) {
        pthread_join(threads[x], NULL);
    }

    free(threads);
}

int main(int argc, char *argv[]) {
    if (argc != 3) {
        printf("usage: proc {s | c} {num of threads}\n");
        exit(1);
    }

    n_threads = atoi(argv[2]);
    latency = (double *) malloc(sizeof(double) * n_threads);
    // printf("Create %d threads, each thread use a socket to execute 1000 times roundtrip\n", n_threads);

    if (*(argv[1]) == 's') {
        server();
    } else {
        client();

        double sum = 0.0;
        for(int x=0; x < n_threads; ++x) {
            sum += latency[x];
            // printf("%2d: %.3f\n", x, latency[x]);
        }
        printf("Avg Latency: %.3f (us)\n", sum / n_threads);
        sleep(1);
    }

    free(latency);
}