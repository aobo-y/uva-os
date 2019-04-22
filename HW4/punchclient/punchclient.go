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

var lclport string // local service port

// Pipe streams between the punch server connection and service connection
// Close both connections when done
func pipe(svrconn net.Conn, lclconn net.Conn) {
	pipeLog := fmt.Sprintf("%v <- %v <--> %v -> %v", svrconn.RemoteAddr().String(), svrconn.LocalAddr().String(), lclconn.LocalAddr().String(), lclconn.RemoteAddr().String())

	fmt.Println("Pipe " + pipeLog)

	// omit channels since client doesn't need to monitor data streamed
	punch.Pipe(svrconn, lclconn, nil, nil)

	svrconn.Close()
	lclconn.Close()

	fmt.Println("Close " + pipeLog)
}

// Spin a new connection to the punch server with the nounce
// Also create a new connection to the backend service
// Pipe the punch server conn & service conn
func connect(svrIp, svrPort, nounce, lclport string) {
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

	lclconn, err := net.Dial("tcp", ":"+lclport)
	if err != nil {
		log.Fatal("Conn to local service error: ", err)
		return
	}

	go pipe(conn, lclconn)
}

// Send request to the punch server to open a port through control conn
// Then wait for connection invites from punch server
// Spin a new connection to each invite
func open(ctrconn net.Conn, srvIp, username, password, rmtport, lclport string) {
	opencmd := "OPEN " + username + " " + password + " " + rmtport

	_, err := ctrconn.Write([]byte(opencmd))
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 1024)
		n, err := ctrconn.Read(buf[:])
		if err != nil {
			panic("Punch server control connection closed")
		}

		msg := string(buf[0:n])

		if msg == "FAIL" {
			panic(fmt.Errorf("Failed to open port in punch server"))
		}

		tokens := strings.Split(msg, " ")
		if len(tokens) == 3 && tokens[0] == "CONNECT" {
			fmt.Println("Connect invite from punch server:", tokens[1])
			go connect(srvIp, tokens[1], tokens[2], lclport)
		} else {
			fmt.Println("Invalid message from punch server:", msg)
		}
	}
}

// Send List request to the punch server through control conn
// Print the server response
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

// Start the punch client and connect to punch server
// Send OPEN or LIST request to punch server
// OPEN port with args of service/outward port, punch server addr, username & password
// LIST with args of punch server addr
func main() {
	var ps string

	if len(os.Args) == 6 || (len(os.Args) == 3 && os.Args[1] == "LIST") {
		ps = os.Args[2]
	} else {
		log.Fatal("Invalid arguments: [local port] [ps addr] [username] [password] [remote port] or LIST [ps addr]")
		return
	}

	ctrconn, err := net.Dial("tcp", ps)

	if err != nil {
		log.Fatal("Dial punch server error: ", err)
		return
	}

	defer ctrconn.Close()
	fmt.Println("Connected to punch server: " + ps)

	if len(os.Args) == 6 {
		lclport := os.Args[1]

		username := os.Args[3]
		password := os.Args[4]
		rmtport := os.Args[5]

		srvIp, _, _ := net.SplitHostPort(ps)

		open(ctrconn, srvIp, username, password, rmtport, lclport)
	} else if len(os.Args) == 3 {
		list(ctrconn)
	}
}
