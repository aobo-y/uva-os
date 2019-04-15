package main

import (
	"fmt"
	"log"
	"net"
)

func echoServer(c net.Conn) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			fmt.Println("Lost connection from " + c.RemoteAddr().String())
			return
		}

		data := buf[0:nr]
		println("Echo Server got:", string(data))

		_, err = c.Write(data)

		if err != nil {
			log.Fatal("Write: ", err)
		}
	}
}

func main() {
	seraddr := "0.0.0.0:8888"

	l, err := net.Listen("tcp", seraddr)

	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Println("Echo server is listening at " + seraddr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		fmt.Println("Received connection from " + conn.RemoteAddr().String())

		go echoServer(conn)
	}
}
