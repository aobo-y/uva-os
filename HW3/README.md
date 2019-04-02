# Machine Problem 3 - FAT32

**Aobo Yang (ay6gv)**

This work is a C++ implementation of the FAT32 file system, which offers interface to mount image, change working directory, list directory, and read files. The interfaces supports both absolute and relative path.

## Problem

This project has four main challenges. First, it should be able to locate the cluster number in the right position of the disk image accoridng to the BPB configuration. Second, it needs to read clusters chained in the FAT. Third, it should be able to parse the path and the directory file strcuture to convert the valid ones to the corresponding cluster numbers. At last, it should maintain the state of the current working directory and file descriptors.

## Implementation

The work exposes the following 6 functions to users.

```c++
extern bool FAT_mount(const char *path);
extern int FAT_cd(const char *path);
extern dirEnt * FAT_readDir(const char *dirname);
extern int FAT_open(const char *path);
extern int FAT_close(int fd);
extern int FAT_pread(int fildes, void *buf, int nbyte, int offset);
```

### Basic Setup

The problem stores several global states to use among different functions. It has to save the mounted image file descriptor for further access. It also keeps the BPB as a `bpbFat32` and the entire FAT table as a `unit32_t` array in memory for the frequent needs of them. `bpbFat32` is accessed for the root cluster number, and number of bytes in a cluster and sector. `FAT` is accessed to chase the chained clauster numbers of a file. The current working directory is saved as a vector of its directory entries. The actual path string is not stored. For the file descriptor, the program create a new structure consisting of three fields: a boolean indicating the avaialability, an integer for file size, and another integer for cluster number. The avaialability boolean can not only help the program find the next free descriptor to open a new file, but also verify if the specified descriptor has a  file opened while closing and reading files. An array of 128 such file descriptors are allocated with availabilities as free. Moreover, a list of the free file descriptor number is created to fast obtain the next available one avoiding tranversing the array.

### FAT_mount




## Analysis

This work can definitely be further optimized. First, keeping the entire BPB in memory is not necessary. Not every field in BPB is useful. The program can calculate and store only the wanted values, like number of bytes in each cluster and the root cluster number. Second, pre-allocate 128-long file descriptor array in the stack may be wasteful. It can be changed to an array of file descriptor pointers and dynamically allocate the actual descriptor in the heap. It can save memories when the number of opened files are usually limited, but this may sacrifies the performance due to the heap operations.

Since the program only keeps the directory entries of the current working directory, it cannot tell the actual absolute path. Even though it is enough for all the fucntions exposed, keep such as path string can be handful to users.
