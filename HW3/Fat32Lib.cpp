#include <iostream>
#include <string>
#include <cstring>
#include <cerrno>
#include <fcntl.h>
#include <unistd.h>
#include <vector>

#include "myfat.h"

int ifd; // image file descriptor
bpbFat32 bpb;
int FirstDataSector;
uint32_t *FAT;
std::vector<dirEnt> cwd_ents;

bool read_sector(int sec_num, int size, void* buf) {
    if (lseek(ifd, sec_num * bpb.bpb_bytesPerSec, SEEK_SET) == -1) {
        return 0;
    }

    std::cout << "set head to " << sec_num * bpb.bpb_bytesPerSec << "\n";
    if (read(ifd, buf, size) == -1) {
        return 0;
    }

    return size;
}

bool read_cluster(int cluster_num, int size, void* buf) {
    int sec_num = FirstDataSector + (cluster_num - 2) * bpb.bpb_secPerClus;

    return read_sector(sec_num, size, buf);
}

uint32_t next_cluster(uint32_t cluster_num) {
    // The high 4 bits of a FAT32 FAT entry are reserved
    return FAT[cluster_num] & 0x0FFFFFFF;
}

std::vector<dirEnt> read_dir_cluster(uint32_t cluster_num) {
    std::vector<uint32_t> clusters;

    do {
        clusters.push_back(cluster_num);
        cluster_num = next_cluster(cluster_num);
    } while (cluster_num < 0x0FFFFFF8);

    int cluster_bytes = bpb.bpb_secPerClus * bpb.bpb_bytesPerSec;
    int dirEnt_count = clusters.size() * cluster_bytes / sizeof(dirEnt);

    std::vector<dirEnt> dirEnt_list (dirEnt_count);

    // convert to bytes array to conduct arithmatic with corresponding size
    uint8_t *ptr = (uint8_t *) dirEnt_list.data();
    for (auto const& cn: clusters) {
        std::cout << cn << " cluster num\n";
        if (!read_cluster(cn, cluster_bytes, ptr)) {
            std::cerr << "Failed to read cluster:" << std::strerror(errno) << "\n";
            break;
        }
        ptr += cluster_bytes;
    }

    return dirEnt_list;
}

bool FAT_mount(const char *path) {
    ifd = open(path, O_RDWR);
    if (ifd == -1) {
        std::cerr << "Failed to open raw image:" << std::strerror(errno) << "\n";
        return 0;
    }

    if (read(ifd, &bpb, sizeof(bpbFat32)) == -1) {
        std::cerr << "Failed to read bdp:" << std::strerror(errno) << "\n";
        return 0;
    }

    std::cout << bpb.bpb_FATSz32<< "\n";

    // RootDirSectors = ((BPB_RootEntCnt * 32) + (BPB_BytsPerSec – 1)) / BPB_BytsPerSec;
    // in FAT32, BPB_RootEntCnt is always 0
    int RootDirSectors = 0;
    int FATSz = bpb.bpb_FATSz32;
    FirstDataSector = bpb.bpb_rsvdSecCnt + (bpb.bpb_numFATs * FATSz) + RootDirSectors;

    int DataSec = bpb.bpb_totSec32 - FirstDataSector;
    int CountofClusters = DataSec / bpb.bpb_secPerClus;
    if (CountofClusters < 65525) {
        std::cerr << "Invalid Fat32 cluster number:" << CountofClusters << "\n";
        return 0;
    }


    std::cout << "cluster count: " << CountofClusters << "\n";
    std::cout << "bpb rsv sec: " << bpb.bpb_rsvdSecCnt << "\n";
    std::cout << "bpb fat size: " << FATSz << "\n";

    int FATBytes = FATSz * bpb.bpb_bytesPerSec;
    FAT = (uint32_t *) malloc(FATBytes);
    if (!read_sector(bpb.bpb_rsvdSecCnt, FATBytes, FAT)) {
        std::cerr << "Failed to read FAT\n";
        return 0;
    }

    // std::cout << "FAT entry count: " << (FATBytes / sizeof(uint32_t)) << "\n";

    std::cout << "Root cluster num: " << bpb.bpb_RootClus << "\n";
    cwd_ents = read_dir_cluster(bpb.bpb_RootClus);
    std::cout << std::string ((char *)cwd_ents[0].dir_name, 11) << "\n";

    return 1;
}


int FAT_cd(const char *path) {
    return 0;
}

int FAT_open(const char *path) {
    return 0;
}

int FAT_close(int fd) {
    return 0;
}

int FAT_pread(int fildes, void *buf, int nbyte, int offset) {
    return 0;
}

dirEnt * FAT_readDir(const char *dirname) {
    return cwd_ents.data();
}

