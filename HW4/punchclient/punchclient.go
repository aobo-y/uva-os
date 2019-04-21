package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"punch"
	"strings"
)

var port string

func pipe(svrconn net.Conn, lclconn net.Conn) {
	fmt.Println("Pipe " + svrconn.RemoteAddr().String() + " <- " + svrconn.LocalAddr().String() + " <--> " + lclconn.LocalAddr().String() + " -> " + lclconn.RemoteAddr().String())

	punch.Pipe(svrconn, lclconn, nil, nil)

	svrconn.Close()
	lclconn.Close()

	fmt.Println("Close " + svrconn.RemoteAddr().String() + " <- " + svrconn.LocalAddr().String() + " <--> " + lclconn.LocalAddr().String() + " -> " + lclconn.RemoteAddr().String())
}

func connect(svrIp string, svrPort string, nounce string) {
	conn, err := net.Dial("tcp", svrIp+":"+svrPort)
	if err != nil {
		log.Fatal("Spin conn to punch server error: ", err)
		return
	}

	defer conn.Close()
	fmt.Println(conn.LocalAddr().String() + " -> " + conn.RemoteAddr().String())

	_, err = conn.Write([]byte(nounce))
	if err != nil {
		log.Fatal("Write nounce to punch server error: ", err)
		return
	}

	lclconn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		log.Fatal("Conn to local service error: ", err)
		return
	}

	go pipe(conn, lclconn)
}

func main() {
	if len(os.Args) < 5 {
		log.Fatal("Invalid arguments: [port] [ps addr] [username] [password]")
		return
	}

	port = os.Args[1]
	ps := os.Args[2]
	username := os.Args[3]
	password := os.Args[4]

	ctrconn, err := net.Dial("tcp", ps)

	if err != nil {
		log.Fatal("Dial punch server error: ", err)
	}

	fmt.Println("Connected to punch server: " + ps)

	opencmd := "OPEN " + username + " " + password + " " + port

	_, err = ctrconn.Write([]byte(opencmd))
	if err != nil {
		log.Fatal("write error: ", err)
		return
	}

	srvIp, _, _ := net.SplitHostPort(ps)

	for {
		buf := make([]byte, 1024)
		n, err := ctrconn.Read(buf[:])
		if err != nil {
			log.Fatal("read error: ", err)
			return
		}

		msg := string(buf[0:n])

		if msg == "FAIL" {
			log.Fatal("Failed to open port in punch server")
			return
		}

		tokens := strings.Split(msg, " ")
		if len(tokens) == 3 && tokens[0] == "CONNECT" {
			fmt.Println("Connect invite from punch server:", tokens[1])
			go connect(srvIp, tokens[1], tokens[2])
		} else {
			fmt.Println("Invalid message from punch server:", msg)
		}
	}
}
