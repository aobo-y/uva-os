msh: libFAT32.so
	gcc -Wall -fPIC -I. -o msh newTest.c -L. -lFAT32

libFAT32.so: Fat32Lib.cpp
	g++ -std=c++11 -shared -o libFAT32.so -fPIC Fat32Lib.cpp

clean: 
	rm *.so msh
