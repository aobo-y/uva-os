package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func readConn(c net.Conn) error {
	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf[:])

		if err != nil {
			return err
		}

		fmt.Println("Echo Client got:", string(buf[0:n]))
	}
}

func writeConn(c net.Conn) error {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := scanner.Text()

		_, err := c.Write([]byte(input))
		if err != nil {
			return err
		}
	}
}

func main() {
	seraddr := "0.0.0.0:8888"

	if len(os.Args) > 1 {
		seraddr = os.Args[1]
	}

	conn, err := net.Dial("tcp", seraddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected to Echo Server " + seraddr + "; Input text below:")

	done := make(chan error)

	// read from echoserver and write to stdout
	go func() {
		done <- readConn(conn)
	}()

	// read from stdin and write to echoserver
	go func() {
		done <- writeConn(conn)
	}()

	err = <-done
	if err != nil && err != io.EOF {
		panic(err)
	}

	fmt.Println("Echo Server closed")
}
