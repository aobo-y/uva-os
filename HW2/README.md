# Machine Promblem 2 - Synchronization

This program is able to find out the max number from `N` input numbers with `N/2` threads in parallel. It allows N to be any integer, not necessarily be power-of-two.

## Problems

This project has three main challenges. First, the jobs should be distributed among several parallel threads. Second, it needs a barrier to ensure the threads work on the results of each other. Third, the intermediate and final results should be shared among threads.

## Implementation

The program starts with a main thread to read and parse stdin, where each line stands for a number. It prints the exception message to stderr when the input is invalid to convert to a number and continue to read next line without crashing. While reaching an empty line, it creates all other worker threads corresponding to the final count of valid numbers received. These worker threads do the actual comparison then, while the main thread only waits them to finsih and print out the  result.

All worker threads are active during the whole comparison process. No thread ends in advance to make sure the below barrier work as expected. In the inital iteration, each thread should have numbers to compare. With the ongoing iteration, more and more threads become idle. The idle threads would directly go to the barrier with nothing to else to do.


### Barrier

The project uses barrier to synchronize the worker threads. The barrier is implemented with semaphore from `<semaphore.h>`. Though it can be a counting semaphore, it is only initialized with either 0 or 1 as a binary semaphore in the program.

The barrier has two integers and three semaphores. The integers are:
- `max_count`: assigned in constructor to indicate the number of threads this barrier should wait
- `count`: begin with 0 to track the number of threads that has arrived; increase whenever a thread come and descrease to 0 when it equals to `max_count`

The semaphores are:
- `mutex`: init with 1; ensure only one thread is accessing `count`
- `waitq`: init with 0; responsible to move the worker threads into the wait queue and wake them up when all threads have arrived
- `handshake`: init with 0; make sure every signal to `waitq` is consumed by a worker thread before triggering another one

Though the barrier provides only one function `wait`, it actually has two different bahaviors based on the `count`. Every incoming thread locks the `mutex` and increase the `count` at first. From here if the `count` is smaller than the `max_count`, the thread releases the `mutex` and go into wait queue and signal the `handshake` when being waken. On the other hand, if `count` equals to `max_count`, the thread keep decreasing the `count` in a loop till zero. In each loop iteration, the thread signals the `waitq` and wait for a `handshake` signal to continue the next one.

The barrier is tested in a scenario that it is inited with `N` as the max number with exact `N` threads. Each thread has a unique id and log of id is added before and after they hit into the barrier. Then repeatedly run the test for thousands times and all the output logs should follow some rules like:
- `N` wait logs should be followed with `N` release logs
- each group of the `N` logs should include each id exactly once

However, even all the outputs are accurate, it cannot prove the implementation is surely correct.

### Intermediate Results

The program uses a vector to store the input number and pass it to all worker threads to act as both the source and result storage.

Every worker thread has a unique identity `idx` from `1` to `N/2`. It is used to assign the initial number indexes, `idx_1` and `idx_2`, of the vector. After each comparison, the thread writes the larger number  back into `idx_1` of the same vector. Then the thread multiply the number indexes by two to compare the largest number of a wider range. For example, in the first first iteration, thread 0 shall compare the number at index 0 and 1, and write the larger number back to 0, while thread 1 shall writes the larger number of index 2 and 3 back to 2. Then in the second iteration, thread 0 shall compare number at 0 and 2, which stands for the largest number in range 0-1 and 2-3, and write the larger one to 0 as well. After the entire vector is covered by thread 0, the largest number should be at index 0.

## Results

The program fulfills the requirements. In addition, it supports decimal input and arbitrary number of inputs without the restrain of power-of-two. Invalid input is also handled.


