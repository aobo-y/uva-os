package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 5 {
		log.Fatal("Invalid arguments: [port] [ps addr] [username] [password]")
		return
	}

	port := os.Args[1]
	ps := os.Args[2]
	username := os.Args[3]
	password := os.Args[4]

	conn, err := net.Dial("tcp", ps)

	if err != nil {
		log.Fatal("Dial punch server "+ps+" error:", err)
	}

	fmt.Println("Connected to punch server:" + ps)

	opencmd := "OPEN " + username + " " + password + " " + port

	_, err = conn.Write([]byte(opencmd))
	if err != nil {
		log.Fatal("write error:", err)
		return
	}

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Fatal("read error:", err)
			return
		}

		msg := string(buf[0:n])

		if msg == "FAIL" {
			log.Fatal("Failed to open port in punch server")
			return
		}
	}
}
