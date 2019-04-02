#include <iostream>
#include <string>
#include <cstring>
#include <cerrno>
#include <fcntl.h>
#include <unistd.h>
#include <vector>
#include <list>

#include "myfat.h"

int ifd; // image file descriptor
bpbFat32 bpb;
int FirstDataSector;
uint32_t *FAT;
std::vector<dirEnt> cwd_ents;

struct fileDesc {
    bool free;
    uint32_t size;
    uint32_t cluster;
};

std::vector<fileDesc> file_descs (128, fileDesc {1, 0, 0});
// a list to avoid searching for free file descriptor
std::list<int> free_fileDescs;

// read bytes into buffer according to a sector
int read_sector(int sec_num, void* buf, int size, int offset) {
    if (lseek(ifd, sec_num * bpb.bpb_bytesPerSec + offset, SEEK_SET) == -1) {
        return -1;
    }

    return read(ifd, buf, size);
}

// read bytes into buffer according to a cluster
int read_cluster(int cluster_num, void* buf, int size, int offset) {
    int sec_num = FirstDataSector + (cluster_num - 2) * bpb.bpb_secPerClus;

    return read_sector(sec_num, buf, size, offset);
}

uint32_t next_cluster(uint32_t cluster_num) {
    // The high 4 bits of a FAT32 FAT entry are reserved
    return FAT[cluster_num] & 0x0FFFFFFF;
}

// recursively read the clusters according to FAT from the given cluster num
// save as vector of dirEnt and return
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
        // std::cout << cn << " cluster num\n";
        if (read_cluster(cn, ptr, cluster_bytes, 0) == -1) {
            std::cerr << "Failed to read cluster:" << std::strerror(errno) << "\n";
            break;
        }
        ptr += cluster_bytes;
    }

    return dirEnt_list;
}

// split path into tokens and use / as root token
std::vector<std::string> tokenize_path(const char *path) {
    std::string s (path);

    std::vector<std::string> tokens;
    size_t pos = 0;
    std::string token;
    std::string delimiter = "/";

    // set root
    if (s.size() > 0 && s[0] == '/') {
        tokens.push_back("/");
    }

    while ((pos = s.find(delimiter)) != std::string::npos) {
        token = s.substr(0, pos);

        // ignore case of aa///bb and leading root
        if (token != "") {
            tokens.push_back(s.substr(0, pos));
        }
        s.erase(0, pos + delimiter.length());
    }

    if (s != "") {
        tokens.push_back(s);
    }

    return tokens;
}

// format the first cluster number of a dirEnt
uint32_t get_cluster_num(dirEnt& ent) {
    return (ent.dir_fstClusHI << 8) | ent.dir_fstClusLO;
}

// return the dirEnt for a given file name in a vector of dirEnts
// return NULL if cannot find the filename
dirEnt* get_dirEnt_by_name(std::vector<dirEnt>& dir_ents, std::string token) {
    std::string name;
    int j;

    for (auto& ent: dir_ents) {
        if (ent.dir_name[0] == 0xE5) {
            continue;
        }
        if (ent.dir_name[0] == 0x00) {
            break;
        }

        name.clear();
        j = 0;
        for(j = 0; j < 11; j++){
            if(ent.dir_name[j] == ' '){
                continue;
            }
            name.push_back(ent.dir_name[j]);
        }

        if (token == name) {
            return &ent;
        }
    }

    return NULL;
}

// return a vector of dirEnts of the path tokens
// return empty vector if path does not exist
std::vector<dirEnt> get_path_dirEnts(std::vector<std::string> path_tokens) {
    std::vector<dirEnt> dir_ents = cwd_ents;
    uint32_t cluster_num;

    for (auto token: path_tokens) {
        if (token == "/") {
            cluster_num = bpb.bpb_RootClus;
        } else {
            dirEnt* entPtr = get_dirEnt_by_name(dir_ents, token);
            // 0x10 stands for ATTR_DIRECTORY
            if (entPtr == NULL || entPtr->dir_attr != 0x10) {
                if (token == "." || token == "..") {
                    // only root has no . .. file
                    // so keep staying in root
                    cluster_num = bpb.bpb_RootClus;
                } else {
                    // path is not a valid directory
                    dir_ents.clear();
                    break;
                }
            } else {
                cluster_num = get_cluster_num(*entPtr);
            }
        }

        if (cluster_num == 0) {
            // 0 means root
            cluster_num = bpb.bpb_RootClus;
        }

        dir_ents = read_dir_cluster(cluster_num);
    }

    return dir_ents;
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

    int FATBytes = FATSz * bpb.bpb_bytesPerSec;
    FAT = (uint32_t *) malloc(FATBytes);
    if (read_sector(bpb.bpb_rsvdSecCnt, FAT, FATBytes, 0) == -1) {
        std::cerr << "Failed to read FAT\n";
        return 0;
    }

    // std::cout << "FAT entry count: " << (FATBytes / sizeof(uint32_t)) << "\n";

    if (FAT_cd("/") == -1) {
        std::cerr << "Failed to set default CWD\n";
    }

    // init free file descriptor list
    for (int i = 0; i < 128; i++) {
        free_fileDescs.push_back(i);
    }

    return 1;
}


int FAT_cd(const char *path) {
    std::vector<std::string> path_tokens = tokenize_path(path);
    std::vector<dirEnt> dir_ents = get_path_dirEnts(path_tokens);

    if (dir_ents.size() == 0) {
        return -1;
    }

    cwd_ents = dir_ents;
    return 1;
}

// create file descriptor for the path and return the fd
// return -1 for failures which including invalid path, not file, no free fd
int FAT_open(const char *path) {
    std::vector<std::string> path_tokens = tokenize_path(path);

    if (path_tokens.size() == 0) {
        return -1;
    }

    std::string name = path_tokens[path_tokens.size() - 1];
    path_tokens.pop_back();

    std::vector<dirEnt> dir_ents = get_path_dirEnts(path_tokens);
    if (dir_ents.size() == 0) {
        return -1;
    }

    dirEnt* entPtr = get_dirEnt_by_name(dir_ents, name);
    // cannot find file or it is a directory
    if (entPtr == NULL || entPtr->dir_attr == 0x10) {
        return -1;
    }

    // no free space
    if (free_fileDescs.size() == 0) {
        return -1;
    }

    int fd = free_fileDescs.front();
    fileDesc fd_ent {0, entPtr->dir_fileSize, get_cluster_num(*entPtr)};
    file_descs[fd] = fd_ent;
    free_fileDescs.pop_front();

    return fd;
}

// close the give file descriptor and return 0
// return -1 if the fd does not exist
int FAT_close(int fd) {
    // make sure file is actually open
    // return -1 if the file descriptor is free
    if (file_descs[fd].free != 0)  {
        return -1;
    }

    file_descs[fd].free = 1;
    free_fileDescs.push_back(fd);

    return 0;
}

// read bytes into the given buf and return the number of bytes read
// return -1 if the fd does not exist
int FAT_pread(int fildes, void *buf, int nbyte, int offset) {
    fileDesc fd_ent = file_descs[fildes];

    if (fd_ent.free != 0) {
        return -1;
    }

    uint32_t cluster_num = fd_ent.cluster;

    // pretend like real file system, always read at least one cluster
    int cluster_bytes = bpb.bpb_secPerClus * bpb.bpb_bytesPerSec;
    uint8_t *ptr = (uint8_t *) buf;
    int left_bytes = nbyte;
    int left_offset = offset;
    int left_filesize = fd_ent.size;

    while (left_bytes != 0 && left_offset < left_filesize) {
        // skip entire cluster if offset is huge
        if (left_offset > cluster_bytes) {
            left_offset -= cluster_bytes;
            left_filesize -= cluster_bytes;
        } else {
            // bytes to read in this cluster
            int read_size = cluster_bytes;

            if (left_offset != 0) {
                read_size -= left_offset;
                left_filesize -= left_offset;
            }

            if (read_size > left_bytes) {
                read_size = left_bytes;
            }

            if (read_size > left_filesize) {
                read_size = left_filesize;
            }

            if (read_cluster(cluster_num, ptr, read_size, left_offset) == -1) {
                return 0;
            }

            ptr += read_size;
            left_bytes -= read_size;
            left_filesize -= read_size;
            left_offset = 0;
        }

        cluster_num = next_cluster(cluster_num);

        // reach file end
        if (cluster_num >= 0x0FFFFFF8) {
            break;
        }
    };

    return nbyte - left_bytes;
}

dirEnt * FAT_readDir(const char *dirname) {
    std::vector<std::string> tokens = tokenize_path(dirname);
    std::vector<dirEnt> dir_ents = get_path_dirEnts(tokens);

    if (dir_ents.size() == 0) {
        return NULL;
    }

    size_t size = sizeof(dirEnt) * dir_ents.size();
    dirEnt *buf = (dirEnt *) malloc(size);
    memcpy(buf, dir_ents.data(), size);

    return buf;
}

