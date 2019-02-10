#include <string>
// #include <cstdlib>
#include <iostream>
#include <vector>
#include <semaphore.h>
#include <pthread.h>

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
                sem_post(&mutex);
                sem_wait(&handshake);
            }

            sem_post(&mutex);
        }
    }
};

struct CompareArgs {
    std::vector<float> *numbers;
    int idx;
};

void *compare(void *param)  {
    CompareArgs *args = (CompareArgs*)param;

    std::cout << "t idx: " << std::to_string(args->idx) + "\n";
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

    int thread_num = numbers.size() / 2;
    pthread_t tids[thread_num];
    pthread_attr_t attr;
    CompareArgs args_list[thread_num];

    pthread_attr_init(&attr);

    for (int i = 0; i < thread_num; i++) {
        args_list[i].numbers = &numbers;
        args_list[i].idx = i;
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
