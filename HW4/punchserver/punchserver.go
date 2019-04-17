package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
)

var optSplit = regexp.MustCompile(`\s+`)

func punch(conn net.Conn, ctrconn net.Conn) {

}

func openPort(ctrconn net.Conn, port string) {
	l, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal("open client port error:", err)
		ctrconn.Write([]byte("FAIL"))
		ctrconn.Close()
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept extrernal connection error:", err)
		}

		go punch(conn, ctrconn)
	}
}

func list() {

}

func verify(uname string, pw string, port string) bool {
	return false
}

func handleClient(conn net.Conn) {
	clistr := conn.RemoteAddr().String()
	fmt.Println("Received connection from " + clistr)

	buf := make([]byte, 512)
	nb, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Lost connection from " + clistr)
		return
	}

	cmd := string(buf[0:nb])

	fmt.Println("Received command from " + clistr + ": " + cmd)

	tokens := optSplit.Split(cmd, -1)

	switch tokens[0] {
	case "OPEN":
		if len(tokens) != 4 {
			fmt.Println("Invalid OPEN operation from " + clistr)
			return
		}

		if !verify(tokens[1], tokens[2], tokens[3]) {
			conn.Write([]byte("FAIL"))
			conn.Close()
			return
		}

		openPort(conn, tokens[3])
	case "LIST":
		if len(tokens) != 1 {
			fmt.Println("Invalid OPEN operation from " + clistr)
			return
		}

		list()
	}
}

func main() {
	port := "9999"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	l, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Println("Punch server is listening at port " + port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go handleClient(conn)
	}
}
