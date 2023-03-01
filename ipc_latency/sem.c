#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include <semaphore.h>
#include <time.h>

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

sem_t sem;
struct timespec t_signal;
struct timespec t_wakeup;

void *wait_thread(void *arg) {
    printf("Waiting for semaphore...\n");
    sem_wait(&sem);
    get_monotonic_time(&t_wakeup);
    printf("Semaphore signalled!\n");

    printf("Semaphore wake-up latency: %ld\n", get_elapsed_time_nano(&t_signal, &t_wakeup));
    return NULL;
}

int main() {
    int i;
    pthread_t threads[THREADS];

    // Initialize the semaphore with value 0
    sem_init(&sem, 0, 0);

    // Create the waiting thread
    pthread_create(&threads[0], NULL, wait_thread, NULL);

    // Sleep for a bit to ensure the waiting thread is blocked on the semaphore
    sleep(1);

    // Signal the semaphore to wake up the waiting thread
    get_monotonic_time(&t_signal);
    sem_post(&sem);

    // Wait for the waiting thread to exit
    for (i = 0; i < THREADS; i++) {
        pthread_join(threads[i], NULL);
    }

    // Destroy the semaphore
    sem_destroy(&sem);

    return 0;
}
