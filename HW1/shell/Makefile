CXX = g++
CXXFLAGS = -Wall -pedantic -O2 -g -std=c++11

all: msh

msh: main.o
	$(CXX) $(CXXFLAGS) -o $@ $^

clean:
	rm -f *.o msh

test: msh
	python3 shell_test.py

archive:
	rm -f shell-archive.tgz
	tar --exclude=shell-archive.tgz --exclude=msh --exclude=.git "--exclude=*.o" -zcvf shell-archive.tgz *

.PHONY: clean all test archive
