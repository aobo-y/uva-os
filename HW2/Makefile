CXX = g++
CXXFLAGS = -Wall -pedantic -O2 -g -std=c++11

all: max

max: main.o
	$(CXX) $(CXXFLAGS) -o $@ $^

clean:
	rm -f *.o max

archive:
	rm -f max-archive.tgz
	tar --exclude=max-archive.tgz --exclude=max --exclude=.git "--exclude=*.o" --exclude=README.md  -zcvf max-archive.tgz *

.PHONY: clean all test archive
