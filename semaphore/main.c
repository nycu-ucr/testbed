#include <stdlib.h>
#include <stdio.h>
#include <semaphore.h>
#include <string.h>
#include <time.h>
#include <pthread.h>
#include <unistd.h>

#define CONNECTION_COUNTS 10

typedef struct list_node Node;
typedef struct linked_list List;
void get_monotonic_time(struct timespec* ts);
long get_time_nano(struct timespec* ts);
long get_elapsed_time_nano(struct timespec* before, struct timespec* after);

struct list_node {
    int id;
    sem_t sem;
    struct timespec ts;
    Node *prev;
    Node *next;
};

struct linked_list {
    Node *head;
    Node *tail;
};

sem_t wq_lock;
List *waiting_queue;
long latency[CONNECTION_COUNTS];

inline void get_monotonic_time(struct timespec* ts) {
    clock_gettime(CLOCK_MONOTONIC, ts);
}

inline long get_time_nano(struct timespec* ts) {
    return (long)ts->tv_sec * 1e9 + ts->tv_nsec;
}

inline long get_elapsed_time_nano(struct timespec* before, struct timespec* after) {
    return get_time_nano(after) - get_time_nano(before);
}

void init() {
    sem_init(&wq_lock, 0, 1);
    waiting_queue = (List *) malloc(sizeof(List));
    waiting_queue->head = NULL;
}

void destroy() {
    sem_destroy(&wq_lock);
    free(waiting_queue);
}

void add_wq(Node *node) {
    sem_wait(&wq_lock);
    if (waiting_queue->head == NULL) {
        // First Node
        waiting_queue->head = node;
        waiting_queue->tail = node;
    } else {
        node->prev = waiting_queue->tail;
        waiting_queue->tail->next = node;
        waiting_queue->tail = node;
    }
    sem_post(&wq_lock);
}

Node *search_wq(int key) {
    Node *result = NULL;
    if (waiting_queue->head != NULL) {
        sem_wait(&wq_lock);
        Node *current = waiting_queue->head;

        if (current->id == key) {
            result = current;
        } else {
            while(current->next != NULL) {
                current = current->next;
                if (current->id == key) {
                    result = current;
                    break;
                }
            }
        }
        sem_post(&wq_lock);
    } else {
        printf("Waiting Queue is empty\n");
    }

    return result;
}

void del_wq(int key) {
    Node *current = search_wq(key);

    sem_wait(&wq_lock);
    if (current->prev != NULL && current->next != NULL) {
        // Middle node
        current->prev->next = current->next;
        current->next->prev = current->prev;
    } else if (current->prev == NULL) {
        // First node
        current->next->prev = NULL;
    } else if (current->next == NULL) {
        // Tail node
        current->prev->next = NULL;
    } else {
        printf("[del_wq]: Error case");
    }
    sem_post(&wq_lock);

    sem_destroy(&current->sem);
    free(current);
}

void simulate_read(int key) {
    Node *node = (Node *) malloc(sizeof(Node));
    struct timespec ts;

    node->id = key;
    node->next = NULL;
    int res = sem_init(&node->sem, 0, 0);
    if (res < 0) {
        perror("read sem init");
    }

    struct timespec t1, t2;
    // get_monotonic_time(&t1);
    add_wq(node);
    // get_monotonic_time(&t2);

    res = sem_wait(&node->sem);
    if (res < 0) {
        perror("[simulate_read]: sem_wait");
    }
    get_monotonic_time(&ts);

    latency[key] = get_elapsed_time_nano(&node->ts, &ts);

    // printf("Add waiting queue time %ld (ns)\n", get_elapsed_time_nano(&t1, &t2));
}

void handler1() {
    Node *node;
    struct timespec t1, t2;
    for(int x=0; x < CONNECTION_COUNTS; ++x) {
        // get_monotonic_time(&t1);
        node = search_wq(x);
        // get_monotonic_time(&t2);
        // printf("Search time %ld (ns)\n", get_elapsed_time_nano(&t1, &t2));

        if (node != NULL) {
            get_monotonic_time(&node->ts);
            int res = sem_post(&node->sem);
            if (res < 0) {
                perror("[handler]: sem_port");
            }
            // printf("Signal %d\n", x);
        } else {
            printf("[handler]: %d not found\n", x);
        }
    }
}

void show_latency() {
    long s = 0;
    for (int x=0; x < CONNECTION_COUNTS; ++x) {
        s += latency[x];
        // printf("%2d: %ld (ns)\n", x, latency[x]);
        if (x == CONNECTION_COUNTS-1) {
            printf("%ld\n", latency[x]);
        } else {
            printf("%ld,", latency[x]);
        }
    }
    printf("\nAvg: %ld (ns)\n", s/CONNECTION_COUNTS);
}

void *reader(void *arg) {
    simulate_read(*((int *)arg));
    pthread_exit(NULL);
}

void *handler(void *arg) {
    handler1();
    pthread_exit(NULL);
}

int main() {
    init();
    pthread_t ps[CONNECTION_COUNTS], p2;

    for (int x=0; x < CONNECTION_COUNTS; ++x) {
        // printf("Create thread: %d\n", x);
        pthread_create(&ps[x], NULL, reader, &x);
        usleep(100);
    }
    sleep(1);
    pthread_create(&p2, NULL, handler, NULL);

    for (int x=0; x < CONNECTION_COUNTS; ++x) {
        pthread_join(ps[x], NULL);
    }
    pthread_join(p2, NULL);

    long s = 0;
    for (int x=0; x < CONNECTION_COUNTS; ++x) {
        s += latency[x];
        // printf("%2d: %ld (ns)\n", x, latency[x]);
        if (x == CONNECTION_COUNTS-1) {
            printf("%ld\n", latency[x]);
        } else {
            printf("%ld,", latency[x]);
        }
    }
    printf("\nAvg: %ld (ns)\n", s/CONNECTION_COUNTS);
}