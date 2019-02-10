#include <string>
// #include <cstdlib>
#include <iostream>
#include <vector>
#include <semaphore.h>

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

    // for (auto f = numbers.begin(); f != numbers.end(); f++) {
    //     std::cout << *f << " ";
    // }
}
