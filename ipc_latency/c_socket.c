#include <time.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define TIMES 20

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

long latency[TIMES];
const u_int16_t port = 50011;

void server() {
    struct sockaddr_in server_address; memset(&server_address, 0, sizeof(struct sockaddr_in));
    int listen_sock = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    ChannelData rxdata;
    struct timespec ts;

    server_address.sin_family = AF_INET;
    server_address.sin_port = htons(port);
    server_address.sin_addr.s_addr = INADDR_ANY;

    int reuse = 1;
    setsockopt(listen_sock, SOL_SOCKET, SO_REUSEADDR, (const char*)&reuse, sizeof(reuse));
    bind(listen_sock, (const struct sockaddr *) &server_address, sizeof(struct sockaddr_in));
    listen(listen_sock, 5);

    int client_sock = accept(listen_sock, NULL, NULL);

    for (int x=0; x<TIMES; ++x) {
        read(client_sock, &rxdata, sizeof(rxdata));

        clock_gettime(CLOCK_MONOTONIC, &ts);
        
        write(client_sock, &rxdata, sizeof(rxdata));
        
        latency[x] = ts.tv_nsec - rxdata.time_;
    }

    close(client_sock);
    close(listen_sock);
}

void client() {
    struct sockaddr_in server_address; memset(&server_address, 0, sizeof(struct sockaddr_in));
    int sock = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    ChannelData txdata;
    struct timespec start, end;

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
        
        latency[x] = end.tv_nsec - start.tv_nsec;  // Round-Trip
    }
    
    close(sock);
}

int main(int argc, char *argv[]) {
    if (*(argv[1]) == 's') {
        server();

    } else {
        client();
    }

    u_int64_t s = 0;
    for(int x=0; x < TIMES; ++x) {
        s += latency[x];
        printf("%2d: %ld (ns)\n", x+1, latency[x]);
    }
    printf("Avg. %ld (ns)\n", s/TIMES);
}