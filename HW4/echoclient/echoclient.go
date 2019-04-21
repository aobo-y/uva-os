package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func reader(c net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf[:])

		if err != nil {
			return
		}

		fmt.Println("Echo Client got:", string(buf[0:n]))
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

	go reader(conn)

	var input string
	for {
		fmt.Scanln(&input)

		_, err := conn.Write([]byte(input))
		if err != nil {
			log.Fatal("write error:", err)
			break
		}
	}
}
