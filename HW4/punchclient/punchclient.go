package main

import (
	"fmt"
	"io/ioutil"
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

	fmt.Println(conn.LocalAddr().String() + " -> " + conn.RemoteAddr().String())

	_, err = conn.Write([]byte(nounce))
	if err != nil {
		log.Fatal("Write nounce to punch server error: ", err)
		return
	}

	// lclconn, err := net.Dial("tcp", ":"+port)
	lclconn, err := net.Dial("tcp", ":8888")
	if err != nil {
		log.Fatal("Conn to local service error: ", err)
		return
	}

	go pipe(conn, lclconn)
}

func open(ctrconn net.Conn, port, srvIp, username, password string) {
	opencmd := "OPEN " + username + " " + password + " " + port

	_, err := ctrconn.Write([]byte(opencmd))
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 1024)
		n, err := ctrconn.Read(buf[:])
		if err != nil {
			panic(err)
		}

		msg := string(buf[0:n])

		if msg == "FAIL" {
			panic(fmt.Errorf("Failed to open port in punch server"))
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

func list(ctrconn net.Conn) {
	_, err := ctrconn.Write([]byte("LIST"))

	if err != nil {
		panic(err)
	}

	buf, err := ioutil.ReadAll(ctrconn)
	if err != nil {
		panic(err)
	}

	fmt.Println("Username Outward_Facing_Port Client_Facing_Port Client_IP Bytes_Received Bytes_Sent")
	fmt.Println(string(buf))
}

func main() {
	var ps string

	if len(os.Args) == 5 || (len(os.Args) == 3 && os.Args[1] == "LIST") {
		ps = os.Args[2]
	} else {
		log.Fatal("Invalid arguments: [port] [ps addr] [username] [password] or LIST [ps addr]")
		return
	}

	ctrconn, err := net.Dial("tcp", ps)

	if err != nil {
		log.Fatal("Dial punch server error: ", err)
		return
	}

	defer ctrconn.Close()
	fmt.Println("Connected to punch server: " + ps)

	if len(os.Args) == 5 {
		port = os.Args[1]

		port = os.Args[1]
		username := os.Args[3]
		password := os.Args[4]

		srvIp, _, _ := net.SplitHostPort(ps)

		open(ctrconn, port, srvIp, username, password)
	} else if len(os.Args) == 3 {
		list(ctrconn)
	}
}
