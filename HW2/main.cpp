#include <string>
#include <iostream>
#include <vector>
#include <semaphore.h>
#include <pthread.h>
#include <math.h>

struct Barrier {
    sem_t mutex;
    sem_t waitq;
    sem_t handshake;
    int count = 0;
    int max_count;

    Barrier(int num) {
        max_count = num;

        sem_init(&mutex, 0, 1);
        sem_init(&waitq, 0, 0);
        sem_init(&handshake, 0, 0);
    }

    void wait() {
        sem_wait(&mutex);

        if (++count < max_count) {
            sem_post(&mutex);

            sem_wait(&waitq);
            sem_post(&handshake);
        } else {
            for (count--; count > 0; count--) {
                sem_post(&waitq);
                sem_wait(&handshake);
            }

            sem_post(&mutex);
        }
    }
};

struct CompareArgs {
    std::vector<float> *numbers;
    int idx;
    Barrier *barrier;
    int iterations;
};

void *compare(void *param)  {
    CompareArgs *args = (CompareArgs*)param;
    std::vector<float> &numbers = *(args->numbers);
    Barrier &barrier = *(args->barrier);

    int idx_1 = 2 * args->idx;
    int idx_2 = idx_1 + 1;

    for (int i = 0; i < args->iterations; i++) {
        // if idx_1 is already out of range, skip
        // when idx_1 in range but idx_2 out, simply skip and leave idx_1 there
        if (idx_2 < numbers.size()) {
            float num_1 = numbers[idx_1];
            float num_2 = numbers[idx_2];

            // write max number to idx_1
            numbers[idx_1] = num_1 > num_2 ? num_1 : num_2;
        }

        barrier.wait();

        // after each iteration, the thread should take care another range
        idx_1 *= 2;
        idx_2 *= 2;
    }

    return 0;
}

int main(void) {
    std::string input;
    std::vector<float> numbers;

    while (true) {
        std::getline(std::cin, input);

        if (input == "") {
            break;
        }

        try {
            float num = std::stof(input);
            numbers.push_back(num);
        } catch (std::exception& e) {
            std::cerr << "invalid input; exception: " << e.what() << '\n';
            continue;
        }
    }

    int thread_num = (numbers.size() + 1) / 2; // ceiling for odd size
    pthread_t tids[thread_num];
    pthread_attr_t attr;
    CompareArgs args_list[thread_num];

    int iterations = ceil(log2(numbers.size()));
    Barrier barrier (thread_num);

    pthread_attr_init(&attr);

    for (int i = 0; i < thread_num; i++) {
        args_list[i].numbers = &numbers;
        args_list[i].idx = i;
        args_list[i].barrier = &barrier;
        args_list[i].iterations = iterations;

        pthread_create(&tids[i], &attr, compare, &args_list[i]);
    }


    for (int i = 0; i < thread_num; i++) {
        pthread_join(tids[i], NULL);
    }

    std::cout << "max number: " << numbers[0] << "\n";

    // for (auto f = numbers.begin(); f != numbers.end(); f++) {
    //     std::cout << *f << " ";
    // }
}
