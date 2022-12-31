#include <stdio.h>
#include <pthread.h>
#include <unistd.h>
#include <time.h>

#define TIMES 1

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
    
} ChannelData;

int pipefd[2];

void *producer(void *arg) {
    FourTuple f = {2130706433, 2130706434, 65530, 65531};
    ChannelData txdata = {1, 2048, NULL, f};
    struct timespec ts;

    for (int x=0; x<TIMES; ++x) {
        clock_gettime(CLOCK_MONOTONIC, &ts);
        printf("%d: Producer write data: %ld (ns)\n", x, ts.tv_nsec);
        write(pipefd[1], &txdata, sizeof(txdata));
    }

    pthread_exit(NULL);
}

void *consumer(void *arg) {
    ChannelData rxdata;
    struct timespec ts;

    for (int x=0; x<TIMES; ++x) {
        read(pipefd[0], &rxdata, sizeof(rxdata));
        clock_gettime(CLOCK_MONOTONIC, &ts);
        printf("%d: Consumer read data: %ld (ns)\n", x, ts.tv_nsec);
    }

    pthread_exit(NULL);
}

int main() {
    pthread_t thread_producer, thread_consumer;

    pipe(pipefd);

    if (pthread_create(&thread_producer, NULL, producer, NULL)) {
        perror("Create thread producer");
    }

    if (pthread_create(&thread_consumer, NULL, consumer, NULL)) {
        perror("Create thread consumer");
    }

    pthread_join(thread_producer, NULL);
    pthread_join(thread_consumer, NULL);

    close(pipefd[0]);
    close(pipefd[1]);
}

/*
345956345 - 345794707 = 161638
697534762 - 697369907 = 164855
697644064 - 697529402 = 114662
697534762 - 697369907 = 164855
697594956 - 697494481 = 100475
485235521 - 485065880 = 169641
796588012 - 796417337 = 170675
92555944 - 92383196   = 172748
*/