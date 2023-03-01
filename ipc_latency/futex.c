#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include <sys/time.h>
#include <sys/syscall.h>
#include <linux/futex.h>
#include <errno.h>
#include <limits.h>
#include <string.h>
#include <stdint.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <time.h>

#define FUTEX_WAIT 0
#define FUTEX_WAKE 1

#define THREADS 2

void get_monotonic_time(struct timespec* ts) {
    clock_gettime(1, ts);
}

long get_time_nano(struct timespec* ts) {
    return (long)ts->tv_sec * 1e9 + ts->tv_nsec;
}

double get_elapsed_time_sec(struct timespec* before, struct timespec* after) {
    double deltat_s  = after->tv_sec - before->tv_sec;
    double deltat_ns = after->tv_nsec - before->tv_nsec;
    return deltat_s + deltat_ns*1e-9;
}

long get_elapsed_time_nano(struct timespec* before, struct timespec* after) {
    return get_time_nano(after) - get_time_nano(before);
}

int futex_wait(volatile int *futex_addr, int val) {
    return syscall(SYS_futex, futex_addr, FUTEX_WAIT, val, NULL, NULL, 0);
}

int futex_wake(volatile int *futex_addr, int val) {
    return syscall(SYS_futex, futex_addr, FUTEX_WAKE, val, NULL, NULL, 0);
}

int futex = 0;
pthread_t threads[THREADS];
struct timespec t_signal;
struct timespec t_wakeup;

void *wait_thread(void *arg) {
    struct timespec t_start;
    int result;

    printf("Waiting for futex...\n");
    result = futex_wait(&futex, 0);
    get_monotonic_time(&t_wakeup);
    printf("Futex woken up!\n");
    if (result == -1) {
        printf("futex_wait error: %s\n", strerror(errno));
        exit(1);
    }

    printf("Futex wake-up latency: %ld\n", get_elapsed_time_nano(&t_signal, &t_wakeup));
    return NULL;
}

int main() {
    int result, i;

    // Initialize the futex to 0
    futex = 0;

    // Create the waiting thread
    result = pthread_create(&threads[0], NULL, wait_thread, NULL);
    if (result) {
        printf("pthread_create error: %s\n", strerror(result));
        exit(1);
    }

    // Sleep for a bit to ensure the waiting thread is blocked on the futex
    sleep(1);

    // Wake up the waiting thread using the futex
    get_monotonic_time(&t_signal);
    result = futex_wake(&futex, 1);
    if (result == -1) {
        printf("futex_wake error: %s\n", strerror(errno));
        exit(1);
    }

    // Wait for the waiting thread to exit
    for (i = 0; i < THREADS; i++) {
        pthread_join(threads[i], NULL);
    }

    return 0;
}
