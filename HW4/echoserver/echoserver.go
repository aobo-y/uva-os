package main

import (
	"fmt"
	"log"
	"net"
)

func echoServer(c net.Conn) {
	clistr := c.RemoteAddr().String()
	fmt.Println("Received connection from " + clistr)

	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			fmt.Println("Lost connection from " + clistr)
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
	port := "8888"

	l, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Println("Echo server is listening at port " + port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go echoServer(conn)
	}
}
