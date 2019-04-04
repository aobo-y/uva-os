# Machine Problem 3 - FAT32

**Aobo Yang (ay6gv)**

This work is a C++ implementation of the FAT32 file system, which offers the interfaces to mount image, change working directory, list directory, and read files. The interfaces supports both absolute and relative path.

## Problem

This project has four main challenges. First, it should be able to locate the cluster number in the right position of the disk image accoridng to the BPB configuration. Second, it needs to read clusters chained in the FAT. Third, it should be able to parse the path and the directory file strcuture to convert the valid ones to the corresponding cluster numbers. At last, it should maintain the state of the current working directory and file descriptors.

## Implementation

The work exposes the following 6 functions to users.

```c++
extern bool FAT_mount(const char *path);
extern dirEnt * FAT_readDir(const char *dirname);
extern int FAT_cd(const char *path);
extern int FAT_open(const char *path);
extern int FAT_close(int fd);
extern int FAT_pread(int fildes, void *buf, int nbyte, int offset);
```

### Basic Setup

The problem stores several global states to use among different functions. It has to save the mounted image file descriptor for further access. It also keeps the BPB as a `bpbFat32` and the entire FAT table as a `unit32_t` array in memory for the frequent needs of them. `bpbFat32` is accessed for the root cluster number, and number of bytes in a cluster and sector. FAT is accessed to chase the chained clauster numbers of a file. The current working directory is saved as a vector of its directory entries. The actual path string is not stored. For the file descriptor, the program create a new structure consisting of three fields: a boolean indicating the avaialability, an integer for file size, and another integer for cluster number. The avaialability boolean can not only help the program find the next free descriptor to open a new file, but also verify if the specified descriptor has a  file opened while closing and reading files. An array of 128 such file descriptors are allocated with availabilities as free. Moreover, a list of the free file descriptor number is created to fast obtain the next available one avoiding tranversing the array.

### Path Parsing

Parsing the path of string format to retreive the corresponding file attributes and location in the disk is the most important functionality that is required in many of the following tasks. This work has implemented several related utility functions serve such needs.

The path is first classified as absolute if it starts with `"/"` or relative otherwise and then tokenized by `"/"`. The starting directory is set to the root for absolute path and current working directory for relative path. For each path token, the program searches for a directory entry in the directory with the same name as the token and then update the directory to the found one. The function iterates over the tokens to find the final entry.

During the process of searching directory entry, the entry name and path token needs to be processed to check equality. According to the FAT specification, the entry name is divided into the main part and the extention part. They are joined with a `"."` after removing the padding space. This produced string is then compared with the uppercased token since FAT stores the entry name in uppercase letters.

A special handling is made for searching for the entry name as `"."` and `".."`. Because the only directory does not has these two is the root, the function ignores any failures of finding such matching entry and stays in the root. It makes more sense that the `"."` can also indicate the current directory even in the root, but it may be weird to see `".."` in the root still leads to the root. However, some other command line interface softwares indeed behave this way. The same handling also applies to the following function searching the path.

### FAT_mount

The function mounts the specified FAT file image, which works as the initialization of the global varibales. The entire mounting process consists of the following steps. First, it opens the file image and save the file descriptor. Then, it reads the BPB and calculates the first data sector to indicate the beginning of clusters and total count of clusters which is used to verify the validity since FAT32 requires more than 65525 clusters. Third, it saves the FAT table by reading the FAT size of bytes immediately after the reserved BPB sectors which are both given in BPB. Forth, the initial current working directory is set to `/`. At last, the linked list of the free file descriptor is initialized with 0 to 127. Any failure among these steps leads to the mount fail.

### FAT_readDir

This function returns an array of the directory entries in the given path. It used the path parsing utility function to tokenize the path and find the final directory. The atribute of the directory entry is used to filter out non-directory files. The content of the last directory is returned as an array of the within entries.


### FAT_cd

This function change the current working directory. It is implemented almost the same as the `FAT_readDir`: tokenizing the path and iterating the token to keep retrieving the next directory. However, instead of returning the array of entries within, it saves the array as the current working dirctory in the global state.

### FAT_open

This function opens the file in the specified path and returns the file descriptor. It first tokenize the path. The last token is removed out as the filename. Then it takes use of `FAT_readDir` with the left tokens to get the parent directory' array of entries. The entry with the exact filename is retrieved. At last, the function pops an empty file descriptor number from the linked list of the free descriptors and saves the entry's file size and the first cluster number into this free position of the file descriptor array. The availablity flag of this  descriptor is marked as false and the number is returned.

### FAT_close

This function is responsible to release the opened file descriptor. It takes a number and check if the availability flag is marked as false in the position of this number within the file desciptor array to ensure the intent of closing an opened file. Then this free flag is updated back as true and the number is also pushed back into the linked list of free descriptors.

### FAT_pread

This function reads the specified number of bytes within the boundary the file size from the offset of the given file descriptor into the buffer. It first ensures the descriptor is opened through the availability flag and gets the file size and the first cluster number. It directly returns if the offset is larger than the file size because there is no contents to read after such offset. Then the function follows the FAT chain to skip clusters to fulfill the offset until the remaining offset is smaller than the cluster size and read the cluster with the offset of the remnant. Each read cannot exceed the boundary of a cluster and the function iterates the FAT chain to read the specified bytes. However, the total bytes read cannot exceed the remaining file size after offset, so the read in the last cluster is smaller than the size occupied by the file even it may not meet the user specified bytes.

## Analysis

This work can definitely be further optimized. First, keeping the entire BPB in memory is not necessary. Not every field in BPB is useful. The program can calculate and store only the wanted values, like number of bytes in each cluster and the root cluster number. Second, pre-allocate 128-long file descriptor array in the stack may be wasteful. It can be changed to an array of file descriptor pointers and dynamically allocate the actual descriptor in the heap. It can save memories when the number of opened files are usually limited, but this may sacrifies the performance due to the heap operations.

Since the program only keeps the directory entries of the current working directory, it cannot tell the actual absolute path. Even though it is enough for all the fucntions exposed, keep such as path string can be handful to users.

The next assigned file descriptor number is not expectable. There is no guarantee that this the smallest available number is returned. Because the available numbers is maintained in a linked list which pop in front and push in the end, the order depends on previously released order.
