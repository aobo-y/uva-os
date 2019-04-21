package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sync"
	"time"

	"punch"

	"golang.org/x/crypto/bcrypt"
)

var users []punch.User

type connection struct {
	sync.RWMutex        // mutex for concurrent read & write
	uname        string // username
	owport       string // outward facing port
	cliport      string // punch client facing port
	cliip        string // client ip address
	bytrcv       int    // bytes received
	bytsnd       int    // bytes sent
}

var openConns map[string]*connection // maps of open connections

var optSplit = regexp.MustCompile(`\s+`)

/*
	Open a
	owconn: outward connection
	ctrconn: control connection
*/
func pipe(owconn net.Conn, cliconn net.Conn) {
	fmt.Println("Pipe " + owconn.RemoteAddr().String() + " -> " + owconn.LocalAddr().String() + " <-->" + cliconn.LocalAddr().String() + " <- " + cliconn.RemoteAddr().String())

	_, port, _ := net.SplitHostPort(owconn.LocalAddr().String())
	connEntry := openConns[port]

	done := make(chan error)
	rcvChn := make(chan int)
	sndChn := make(chan int)

	go func() {
		done <- punch.Pipe(owconn, cliconn, rcvChn, sndChn)
	}()

	select {
	case n := <-rcvChn:
		connEntry.Lock()
		connEntry.bytrcv += n
		connEntry.Unlock()

	case n := <-sndChn:
		connEntry.Lock()
		connEntry.bytsnd += n
		connEntry.Unlock()

	case <-done:
		break
	}

	owconn.Close()
	cliconn.Close()

	fmt.Println("Close " + owconn.RemoteAddr().String() + " -> " + owconn.LocalAddr().String() + " <-->" + cliconn.LocalAddr().String() + " <- " + cliconn.RemoteAddr().String())
}

func listen(ctrconn net.Conn, port string, uname string) {
	l, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal("open outward facing port error: ", err)
		ctrconn.Write([]byte("FAIL"))
		ctrconn.Close()
		return
	}

	defer l.Close()

	clilstner, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("open punch client port error: ", err)
		return
	}

	defer clilstner.Close()

	_, cliport, _ := net.SplitHostPort(clilstner.Addr().String())

	cliip, _, _ := net.SplitHostPort(ctrconn.RemoteAddr().String())
	openConns[port] = &connection{
		uname:   uname,
		owport:  port,
		cliport: cliport,
		cliip:   cliip,
	}
	defer delete(openConns, port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept extrernal connection error: ", err)
		}

		fmt.Println("Outward " + conn.RemoteAddr().String() + " -> " + conn.LocalAddr().String())

		nounce := randString(256)
		ctrconn.Write([]byte("CONNECT " + cliport + " " + nounce))

		// thread to wait for connection with correct nounce
		connChn := make(chan net.Conn)
		errChn := make(chan error)
		go func() {
			cliconn, err := clilstner.Accept()
			if err != nil {
				errChn <- fmt.Errorf("accept punch client connection error: %v", err)
				return
			}

			buf := make([]byte, 512)
			cliconn.Read(buf)
			clinounce := string(buf)

			if clinounce != nounce {
				errChn <- fmt.Errorf("Punch client connection fail, received: %v", clinounce)
				return
			}

			connChn <- cliconn
		}()

		// timeout after 10s
		select {
		case cliconn := <-connChn:
			fmt.Println("Client " + cliconn.RemoteAddr().String() + " -> " + cliconn.LocalAddr().String())
			go pipe(conn, cliconn)
		case err = <-errChn:
			fmt.Println(err)
			return
		case <-time.After(10 * time.Second):
			fmt.Println("Punch client connection timeout")
			return
		}
	}
}

func list() {

}

func verify(uname string, pw string, port string) bool {
	pwbyt := []byte(pw)

	for _, user := range users {
		if user.Uname != uname {
			continue
		}

		if user.Port != port {
			return false
		}

		hpswd := []byte(user.Pswd)
		err := bcrypt.CompareHashAndPassword(hpswd, pwbyt)

		if err != nil {
			return false
		}

		return true
	}

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

		go openPort(conn, tokens[3], tokens[1])
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
		log.Fatal("listen error: ", err)
	}

	fmt.Println("Punch server is listening at port " + port)

	users = *punch.ReadUsers()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error: ", err)
		}

		go handleClient(conn)
	}
}
