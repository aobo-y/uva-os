#include <iostream>
#include <string>
#include <cstring>
#include <cerrno>
#include <fcntl.h>

#include 'myfat.h'

int ifd; // image file descriptor
bpbFat32 bpb;

bool FAT_mount(const char *path) {
    fd = open(path, O_RDWR);
    if (fd == -1) {
        std::cerr << "Failed to open raw image:" << std::strerror(errno) << "\n";
        return -1;
    }

    if (read(ifd, bpb, sizeof(bpbFat32) == -1) {
        std::cerr << "Failed to read bdp:" << std:strerror(errno) << "\n";
        return -1;
    }

    std::cout << bpb->bpb_numFATs << "\n";
}


